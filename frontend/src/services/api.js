// src/services/api.js
import axios from 'axios';

const baseUrl = process.env.BASEURL;

const api = axios.create({
    baseUrl, // A URL base do nosso backend
});

export default api;