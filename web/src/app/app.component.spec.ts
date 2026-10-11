import { TestBed } from '@angular/core/testing';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiService } from './shared/api.service';
import { AppComponent } from './app.component';

describe('Integración de sesión de AppComponent', () => {
  beforeEach(() => {
    sessionStorage.clear();
    localStorage.clear();
  });

  it('cambia de autenticación a cálculo y vuelve al cerrar sesión', async () => {
    await TestBed.configureTestingModule({
      imports: [AppComponent],
      providers: [{
        provide: ApiService,
        useValue: { checkHealth: vi.fn(), factorize: vi.fn(), login: vi.fn(), errorMessage: vi.fn() },
      }],
    }).compileComponents();
    const fixture = TestBed.createComponent(AppComponent);
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('app-auth-page')).not.toBeNull();

    fixture.componentInstance.onAuthenticated({
      accessToken: 'signed-token', tokenType: 'Bearer', expiresIn: 900, expiresAt: new Date(Date.now() + 900_000).toISOString(),
    });
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('app-calculation-page')).not.toBeNull();
    expect(sessionStorage.getItem('matrix-lab.access-token')).toBe('signed-token');

    fixture.componentInstance.logout();
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('app-auth-page')).not.toBeNull();
    expect(sessionStorage.getItem('matrix-lab.access-token')).toBeNull();
  });
});
