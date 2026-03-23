let mocha = require('mocha');
let assert = require('assert');
let scenario2test_login_fixture = require('..');
describe('test suite', function() {
    it('test case', function(done) {
        const assert = require('assert');
describe('processLogin', function () {
  it('returns invalid credentials for wrong password', function () {
    const mod = require('./index');
    const result = mod.processLogin({ username: 'valid_user', password: 'invalid_password' });
    assert.equal(result.ok, false);
    assert.equal(result.message, 'Invalid credentials');
  });
});
    })
})