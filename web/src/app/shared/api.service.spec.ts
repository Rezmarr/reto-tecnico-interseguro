import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { ApiService } from './api.service';
import { FactorizationResponse } from './models';

describe('ApiService HTTP integration', () => {
  let api: ApiService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({ providers: [provideHttpClient(), provideHttpClientTesting()] });
    api = TestBed.inject(ApiService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('requests and returns a token from the Go authentication route', async () => {
    const response = { accessToken: 'signed-token', tokenType: 'Bearer' as const, expiresIn: 900, expiresAt: '2030-01-01T00:15:00Z' };
    const promise = api.login('https://go.example.test', 'demo', 'password');
    const request = http.expectOne('https://go.example.test/api/v1/auth/token');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({ username: 'demo', password: 'password' });
    request.flush(response);
    await expect(promise).resolves.toEqual(response);
  });

  it('attaches the user token to the factorization request', async () => {
    const response = { shape: { rows: 1, columns: 1 }, q: [[1]], r: [[2]], statistics: {} } as unknown as FactorizationResponse;
    const promise = api.factorize('https://go.example.test', [[2]], 'signed-token');
    const request = http.expectOne('https://go.example.test/api/v1/factorizations');
    expect(request.request.method).toBe('POST');
    expect(request.request.headers.get('Authorization')).toBe('Bearer signed-token');
    expect(request.request.body).toEqual({ matrix: [[2]] });
    request.flush(response);
    await expect(promise).resolves.toEqual(response);
  });
});
