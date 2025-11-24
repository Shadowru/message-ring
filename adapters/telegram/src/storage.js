const fs = require('fs');
const path = require('path');

const STORE_PATH = '/app/store/sessions.json';

// Инициализация файла, если нет
if (!fs.existsSync(STORE_PATH)) {
    if (!fs.existsSync(path.dirname(STORE_PATH))) {
        fs.mkdirSync(path.dirname(STORE_PATH), { recursive: true });
    }
    fs.writeFileSync(STORE_PATH, JSON.stringify({}));
}

function getAllSessions() {
    try {
        return JSON.parse(fs.readFileSync(STORE_PATH, 'utf8'));
    } catch (e) {
        return {};
    }
}

function saveSession(sessionId, sessionString) {
    const data = getAllSessions();
    data[sessionId] = sessionString;
    fs.writeFileSync(STORE_PATH, JSON.stringify(data, null, 2));
}

function getSession(sessionId) {
    const data = getAllSessions();
    return data[sessionId];
}

module.exports = { getAllSessions, saveSession, getSession };