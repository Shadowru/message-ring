const express = require('express');
const bodyParser = require('body-parser');
const manager = require('./manager');

const app = express();
app.use(bodyParser.json());

const PORT = 3000;

// 1. Инициализация входа (запрос от Core)
app.post('/connect', async (req, res) => {
    const { sessionId } = req.body; // Уникальный ID аккаунта из БД Core
    if (!sessionId) return res.status(400).json({ error: 'sessionId required' });

    try {
        const result = await manager.loginWithQR(sessionId);
        res.json(result);
    } catch (e) {
        res.status(500).json({ error: e.message });
    }
});

// 2. Получение QR кода (поллинг с фронтенда)
app.get('/qr/:sessionId', (req, res) => {
    const { sessionId } = req.params;
    
    // Проверяем, авторизован ли уже
    if (manager.activeClients[sessionId]) {
        // Тут нужна проверка checkAuthorization, но для скорости проверим наличие QR
        // Если QR нет, а клиент есть - скорее всего подключен
    }

    const qr = manager.pendingQRs[sessionId];
    if (qr) {
        res.json({ status: 'waiting_scan', qr: qr });
    } else {
        // Проверяем, может уже вошли
        const client = manager.activeClients[sessionId];
        if (client) {
             // В идеале сделать асинхронную проверку, но пока так
            res.json({ status: 'connected', qr: null });
        } else {
            res.status(404).json({ status: 'not_found' });
        }
    }
});

// 3. Статус
app.get('/status', (req, res) => {
    res.json({ 
        active_sessions: Object.keys(manager.activeClients).length 
    });
});

// Запуск
app.listen(PORT, async () => {
    console.log(`TG Adapter running on port ${PORT}`);
    await manager.restoreSessions();
});