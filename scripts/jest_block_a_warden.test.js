const request = require('supertest');

const BASE_URL = process.env.API_URL || 'http://localhost:3000';
const wardenCredentials = {
  email: 'admin1@uni.com',
  password: 'admin1',
};

describe('Block A Warden Login and Complaint Fetch', () => {
  let token;

  it('should log in as Block A warden', async () => {
    const res = await request(BASE_URL)
      .post('/api/login')
      .send(wardenCredentials)
      .set('Accept', 'application/json');
    expect(res.status).toBe(200);
    expect(res.body).toHaveProperty('token');
    token = res.body.token;
  });

  it('should fetch complaints as Block A warden', async () => {
    expect(token).toBeDefined();
    const res = await request(BASE_URL)
      .get('/api/complaints')
      .set('Authorization', `Bearer ${token}`);
    expect(res.status).toBe(200);
    expect(res.body).toHaveProperty('count');
    expect(Array.isArray(res.body.data)).toBe(true);
  });
});
