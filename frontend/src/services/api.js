// src/services/api.js
import axios from 'axios';

const api = axios.create({
    baseURL: '/api', // A URL base do nosso backend
});

export default api;