// src/services/api.js
import axios from 'axios';

const api = axios.create({
    baseURL: 'https://fifo-system.onrender.com:8080', // A URL base do nosso backend
});

export default api;