import axios from 'axios';

var base = 'http://localhost:8078/api/v1/';

export const ImageApi = axios.create({
  baseURL: base + 'image',
  timeout: 60000
});
