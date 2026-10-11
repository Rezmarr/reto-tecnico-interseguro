import { TestBed } from '@angular/core/testing';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiService } from '../../shared/api.service';
import { TokenResponse } from '../../shared/models';
import { AuthPageComponent } from './auth-page.component';

describe('Componente de autenticación', () => {
  let api: { login: ReturnType<typeof vi.fn>; errorMessage: ReturnType<typeof vi.fn> };

  beforeEach(async () => {
    api = {
      login: vi.fn().mockResolvedValue({ accessToken: 'signed-token', tokenType: 'Bearer', expiresIn: 900, expiresAt: '2030-01-01T00:15:00Z' } satisfies TokenResponse),
      errorMessage: vi.fn().mockReturnValue({ message: 'Error de API' }),
    };
    await TestBed.configureTestingModule({
      imports: [AuthPageComponent],
      providers: [{ provide: ApiService, useValue: api }],
    }).compileComponents();
  });

  it('emite el token generado y guarda la URL normalizada de Go', async () => {
    localStorage.removeItem('matrix-lab.go-api-url');
    const fixture = TestBed.createComponent(AuthPageComponent);
    const page = fixture.componentInstance;
    const authenticated = vi.fn();
    page.authenticated.subscribe(authenticated);
    page.username.set('demo');
    page.password.set('secret');
    await page.login();
    expect(api.login).toHaveBeenCalledWith('http://localhost:8080', 'demo', 'secret');
    expect(authenticated).toHaveBeenCalledWith(expect.objectContaining({ accessToken: 'signed-token' }));
    expect(localStorage.getItem('matrix-lab.go-api-url')).toBe('http://localhost:8080');
  });

  it('no llama a la API cuando la URL de Go no es válida', async () => {
    const fixture = TestBed.createComponent(AuthPageComponent);
    fixture.componentInstance.goUrl.set('ftp://invalid');
    await fixture.componentInstance.login();
    expect(api.login).not.toHaveBeenCalled();
    expect(fixture.componentInstance.error()).toMatch(/dirección/i);
  });
});
