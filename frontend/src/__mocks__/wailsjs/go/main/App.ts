import { vi } from 'vitest';

export const Greet = vi.fn().mockImplementation((name: string) => {
  if (name === '') {
    return Promise.resolve("Hello, it's nice to meet you!");
  }
  return Promise.resolve(`Hello ${name}, welcome to SMTP-EDC!`);
});
