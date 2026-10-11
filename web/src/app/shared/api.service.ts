import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom, timeout } from 'rxjs';
import { ApiError, FactorizationResponse, TokenResponse } from './models';

@Injectable({ providedIn: 'root' })
export class ApiService {
  private readonly http = inject(HttpClient);

  async checkHealth(baseUrl: string): Promise<void> {
    await firstValueFrom(this.http.get(`${baseUrl}/health`).pipe(timeout({ first: 90_000 })));
  }

  async login(baseUrl: string, username: string, password: string): Promise<TokenResponse> {
    return firstValueFrom(this.http.post<TokenResponse>(`${baseUrl}/api/v1/auth/token`, { username, password }).pipe(timeout({ first: 15_000 })));
  }

  async factorize(baseUrl: string, matrix: number[][], token: string): Promise<FactorizationResponse> {
    return firstValueFrom(
      this.http.post<FactorizationResponse>(`${baseUrl}/api/v1/factorizations`, { matrix }, { headers: { Authorization: `Bearer ${token}` } })
        .pipe(timeout({ first: 105_000 })),
    );
  }

  errorMessage(error: unknown): { message: string; requestId?: string } {
    if (error instanceof HttpErrorResponse) {
      const body = error.error as ApiError | null;
      return {
        message: body?.error?.message ?? this.statusMessage(error.status),
        requestId: body?.error?.requestId ?? error.headers.get('X-Request-ID') ?? undefined,
      };
    }
    if (error instanceof Error && error.name === 'TimeoutError') {
      return { message: 'La solicitud tardó más de lo esperado. Comprueba que las API estén disponibles e inténtalo de nuevo.' };
    }
    return { message: 'No se pudo conectar con la API. Revisa las direcciones configuradas y vuelve a intentarlo.' };
  }

  private statusMessage(status: number): string {
    if (status === 0) return 'No se pudo conectar con la API. Puede estar iniciando o la dirección no ser correcta.';
    if (status === 401) return 'La API rechazó la solicitud.';
    if (status === 413) return 'La matriz supera el tamaño permitido.';
    return `La API respondió con un error HTTP ${status}.`;
  }
}
