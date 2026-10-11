import { Component, EventEmitter, Output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { environment } from '../../../environments/environment';
import { ApiService } from '../../shared/api.service';
import { TokenResponse } from '../../shared/models';

const STORAGE_GO_URL = 'matrix-lab.go-api-url';

@Component({
  selector: 'app-auth-page',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './auth-page.component.html',
  styleUrl: './auth-page.component.css',
})
export class AuthPageComponent {
  @Output() authenticated = new EventEmitter<TokenResponse>();

  readonly username = signal('');
  readonly password = signal('');
  readonly goUrl = signal(this.readGoUrl());
  readonly error = signal('');
  readonly busy = signal(false);

  constructor(private readonly api: ApiService) {}

  async login(): Promise<void> {
    this.error.set('');
    const goUrl = this.normalizeUrl(this.goUrl());
    if (!goUrl) {
      this.error.set('Ingresa una dirección HTTP o HTTPS válida para la API Go.');
      return;
    }

    this.busy.set(true);
    try {
      const response = await this.api.login(goUrl, this.username(), this.password());
      localStorage.setItem(STORAGE_GO_URL, goUrl);
      this.password.set('');
      this.authenticated.emit(response);
    } catch (error) {
      const message = this.api.errorMessage(error).message;
      this.error.set(message === 'La API rechazó la solicitud.' ? 'Usuario o contraseña incorrectos.' : message);
    } finally {
      this.busy.set(false);
    }
  }

  private readGoUrl(): string {
    if (typeof localStorage === 'undefined') return environment.goApiUrl;
    return localStorage.getItem(STORAGE_GO_URL) ?? environment.goApiUrl;
  }

  private normalizeUrl(value: string): string | null {
    const trimmed = value.trim();
    if (trimmed.startsWith('/') && !trimmed.startsWith('//') && !trimmed.includes('?') && !trimmed.includes('#')) {
      return trimmed.replace(/\/+$/, '') || '/';
    }
    try {
      const url = new URL(trimmed);
      if (url.protocol !== 'http:' && url.protocol !== 'https:') return null;
      return url.toString().replace(/\/+$/, '');
    } catch {
      return null;
    }
  }
}
