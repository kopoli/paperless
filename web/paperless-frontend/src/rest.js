import axios from 'axios';

var base = '/api/v1/';

export const ImageApi = axios.create({
  baseURL: base + 'image',
  timeout: 60000
});

export const TagApi = axios.create({
  baseURL: base + 'tag',
  timeout: 60000
});
