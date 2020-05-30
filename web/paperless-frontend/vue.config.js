module.exports = {
  publicPath: process.env.NODE_ENV === 'production'
    ? '/dist/'
    : '/',
  devServer: {
    proxy: {
      '/api/v1': {
        target: 'http://localhost:8078',
        secure: false
      },
      '/static/': {
        target: 'http://localhost:8078',
        secure: false
      }
    }
  }
};
