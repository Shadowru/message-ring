const { TelegramClient } = require("telegram");
const { StringSession } = require("telegram/sessions");
const { NewMessage } = require("telegram/events");
const QRCode = require("qrcode");
const axios = require("axios");
const storage = require("./storage");

const API_ID = parseInt(process.env.TG_API_ID);
const API_HASH = process.env.TG_API_HASH;
const CORE_WEBHOOK_URL = process.env.CORE_WEBHOOK_URL;

// Храним активные клиенты в памяти: { sessionId: TelegramClient }
const activeClients = {};
// Храним текущие QR коды для логина: { sessionId: "base64_image..." }
const pendingQRs = {};

/**
 * Запуск клиента (восстановление или новый вход)
 */
async function startClient(sessionId, savedSessionString = "") {
    if (activeClients[sessionId]) return activeClients[sessionId];

    const stringSession = new StringSession(savedSessionString);
    
    const client = new TelegramClient(stringSession, API_ID, API_HASH, {
        connectionRetries: 5,
        useWSS: false, // Используем TCP
    });

    // Подключаемся к серверам TG
    await client.connect();

    // Вешаем обработчик входящих сообщений
    client.addEventHandler(async (event) => {
        const message = event.message;
        if (message.isPrivate) { // Обрабатываем только личные сообщения (можно настроить)
            await sendToCore(sessionId, message);
        }
    }, new NewMessage({ incoming: true }));

    activeClients[sessionId] = client;
    return client;
}

/**
 * Процесс входа по QR
 */
async function loginWithQR(sessionId) {
    const client = await startClient(sessionId, ""); // Запускаем чистый клиент

    // Если уже авторизован (например, сессия не пустая)
    if (await client.checkAuthorization()) {
        return { status: "connected" };
    }

    // Запускаем флоу QR кода
    // client.signInWithQrCode возвращает промис, который резолвится ТОЛЬКО после сканирования
    // Поэтому мы не ждем его через await, а вешаем колбэки
    client.signInWithQrCode({ apiId: API_ID, apiHash: API_HASH }, {
        onError: (err) => console.error("QR Error:", err),
        qrCode: async (code) => {
            // Генерируем картинку Base64 для фронтенда
            const qrImage = await QRCode.toDataURL(code.token.toString('base64'));
            pendingQRs[sessionId] = qrImage;
            console.log(`QR generated for ${sessionId}`);
        }
    }).then(async () => {
        // Успешный вход!
        console.log(`User logged in: ${sessionId}`);
        const sessionString = client.session.save();
        storage.saveSession(sessionId, sessionString);
        delete pendingQRs[sessionId];
        
        // Можно уведомить Core, что статус изменился
    }).catch(err => {
        console.error("Login failed:", err);
    });

    return { status: "qr_generated" };
}

/**
 * Отправка вебхука в Core
 */
async function sendToCore(sessionId, tgMessage) {
    try {
        // Получаем инфо о сендере
        const sender = await tgMessage.getSender();
        
        const payload = {
            sessionId: sessionId,
            event: "message",
            data: {
                id: tgMessage.id,
                text: tgMessage.text,
                date: tgMessage.date,
                senderId: sender.id.toString(),
                senderName: sender.firstName + (sender.lastName ? " " + sender.lastName : ""),
                senderUsername: sender.username,
                senderPhone: sender.phone,
                // TODO: Добавить обработку медиа (скачивание)
            }
        };

        console.log(`Forwarding message from ${sender.firstName} to Core...`);
        await axios.post(CORE_WEBHOOK_URL, payload);
    } catch (e) {
        console.error("Failed to send webhook to Core:", e.message);
    }
}

// Восстановление сессий при старте контейнера
async function restoreSessions() {
    const sessions = storage.getAllSessions();
    for (const [id, sessionStr] of Object.entries(sessions)) {
        console.log(`Restoring session: ${id}`);
        try {
            const client = await startClient(id, sessionStr);
            if (!await client.checkAuthorization()) {
                console.warn(`Session ${id} invalid/expired`);
            }
        } catch (e) {
            console.error(`Failed to restore ${id}:`, e);
        }
    }
}

module.exports = { loginWithQR, restoreSessions, pendingQRs, activeClients };