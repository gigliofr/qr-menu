module.exports = {
  timeout: 30000,
  use: {
    headless: true,
    baseURL: process.env.BASE_URL || 'http://localhost:8080',
  },
  testDir: './',
};
