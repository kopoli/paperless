import axios from 'axios';

var base = '/api/v1/';

export const ImageApi = axios.create({
  baseURL: base + 'image',
  timeout: 60000
});
