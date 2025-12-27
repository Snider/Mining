export const API_BASE = 'http://localhost:9090/api/v1/mining';

// Test XMR wallet address
export const TEST_XMR_WALLET = '89qpYgfAZzp8VYKaPbAh1V2vSW9RHCMyHVQxe2oFxZvpK9dF1UMpZSxJK9jikW4QCRGgVni8BJjvTQpJQtHJzYyw8Uz18An';

// Test mining pool
export const TEST_POOL = 'pool.supportxmr.com:3333';

export const testProfile = {
  name: 'Test Profile',
  minerType: 'xmrig',
  config: {
    pool: TEST_POOL,
    wallet: TEST_XMR_WALLET,
    tls: false,
    hugePages: true,
  },
};

export const testProfileMinimal = {
  name: 'Minimal Test Profile',
  minerType: 'xmrig',
  config: {
    pool: TEST_POOL,
    wallet: TEST_XMR_WALLET,
  },
};
