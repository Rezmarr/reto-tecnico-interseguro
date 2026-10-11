import { Component, signal } from '@angular/core';
import { AuthPageComponent } from './features/auth/auth-page.component';
import { CalculationPageComponent } from './features/calculation/calculation-page.component';
import { TokenResponse } from './shared/models';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [AuthPageComponent, CalculationPageComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent {
  readonly accessToken = signal(this.loadAccessToken());

  onAuthenticated(response: TokenResponse): void {
    sessionStorage.setItem('matrix-lab.access-token', response.accessToken);
    sessionStorage.setItem('matrix-lab.token-expires-at', response.expiresAt);
    this.accessToken.set(response.accessToken);
  }

  logout(): void {
    sessionStorage.removeItem('matrix-lab.access-token');
    sessionStorage.removeItem('matrix-lab.token-expires-at');
    this.accessToken.set('');
  }

  private loadAccessToken(): string {
    if (typeof sessionStorage === 'undefined') return '';
    const token = sessionStorage.getItem('matrix-lab.access-token') ?? '';
    const expiresAt = sessionStorage.getItem('matrix-lab.token-expires-at') ?? '';
    if (!token || !expiresAt || Date.parse(expiresAt) <= Date.now()) {
      sessionStorage.removeItem('matrix-lab.access-token');
      sessionStorage.removeItem('matrix-lab.token-expires-at');
      return '';
    }
    return token;
  }
}
