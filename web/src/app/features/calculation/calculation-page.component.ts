import { Component, EventEmitter, Input, Output, signal, WritableSignal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { environment } from '../../../environments/environment';
import { ApiService } from '../../shared/api.service';
import { FactorizationResponse, ServiceState, ServiceStatus } from '../../shared/models';

const STORAGE_GO_URL = 'matrix-lab.go-api-url';
const STORAGE_NODE_URL = 'matrix-lab.node-api-url';

@Component({
  selector: 'app-calculation-page',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './calculation-page.component.html',
  styleUrl: './calculation-page.component.css',
})
export class CalculationPageComponent {
  @Input({ required: true }) accessToken = '';
  @Output() logoutRequested = new EventEmitter<void>();
  readonly rowCount = signal(3);
  readonly columnCount = signal(2);
  readonly matrixCells = signal<string[][]>([['1', '0'], ['0', '1'], ['1', '1']]);
  readonly goUrl = signal(this.readSetting(STORAGE_GO_URL, environment.goApiUrl));
  readonly nodeUrl = signal(this.readSetting(STORAGE_NODE_URL, environment.nodeApiUrl));
  readonly result = signal<FactorizationResponse | null>(null);
  readonly busy = signal(false);
  readonly settingsOpen = signal(false);
  readonly phase = signal<'idle' | 'checking' | 'calculating'>('idle');
  readonly inputError = signal('');
  readonly requestError = signal('');
  readonly requestId = signal('');
  readonly coldStartNotice = signal(false);
  readonly goStatus = signal<ServiceStatus>({ state: 'idle', elapsed: 0 });
  readonly nodeStatus = signal<ServiceStatus>({ state: 'idle', elapsed: 0 });

  private readonly presets: Record<string, string[][]> = {
    identity: [['1', '0'], ['0', '1']],
    tall: [['1', '0'], ['0', '1'], ['1', '1']],
    wide: [['1', '2', '3'], ['4', '5', '6']],
  };

  constructor(private readonly api: ApiService) {}

  logout(): void { this.logoutRequested.emit(); }

  get matrixPairs(): Array<{ label: string; name: string; values: number[][] }> {
    const current = this.result();
    return current ? [
      { label: 'Q', name: 'Matriz ortogonal', values: current.q },
      { label: 'R', name: 'Matriz triangular', values: current.r },
    ] : [];
  }

  get columnIndexes(): number[] {
    return Array.from({ length: this.columnCount() }, (_, index) => index);
  }

  usePreset(name: keyof CalculationPageComponent['presets']): void {
    const preset = this.presets[name].map((row) => [...row]);
    this.rowCount.set(preset.length);
    this.columnCount.set(preset[0].length);
    this.matrixCells.set(preset);
    this.inputError.set('');
    this.requestError.set('');
    this.result.set(null);
  }

  resizeRows(value: number | string | null): void { this.resizeMatrix(Number(value), this.columnCount()); }
  resizeColumns(value: number | string | null): void { this.resizeMatrix(this.rowCount(), Number(value)); }

  setCellValue(rowIndex: number, columnIndex: number, value: string): void {
    const next = this.matrixCells().map((row) => [...row]);
    next[rowIndex][columnIndex] = value;
    this.matrixCells.set(next);
    this.inputError.set('');
    this.result.set(null);
  }

  saveConnections(): void {
    const go = this.normalizeUrl(this.goUrl());
    const node = this.normalizeUrl(this.nodeUrl());
    if (!go || !node) {
      this.requestError.set('Ingresa una dirección válida para las API de Go y Node.');
      return;
    }
    this.goUrl.set(go);
    this.nodeUrl.set(node);
    localStorage.setItem(STORAGE_GO_URL, go);
    localStorage.setItem(STORAGE_NODE_URL, node);
    this.settingsOpen.set(false);
    this.requestError.set('');
  }

  toggleSettings(): void {
    this.settingsOpen.update((open) => !open);
  }

  async factorize(): Promise<void> {
    this.inputError.set('');
    this.requestError.set('');
    this.requestId.set('');
    let matrix: number[][];
    try {
      matrix = this.parseMatrix(this.matrixCells());
    } catch (error) {
      this.inputError.set(error instanceof Error ? error.message : 'La matriz no tiene un formato válido.');
      return;
    }
    const goUrl = this.normalizeUrl(this.goUrl());
    const nodeUrl = this.normalizeUrl(this.nodeUrl());
    if (!goUrl || !nodeUrl) {
      this.requestError.set('Configura las direcciones HTTP o HTTPS de las API antes de calcular.');
      this.settingsOpen.set(true);
      return;
    }

    this.busy.set(true);
    this.phase.set('checking');
    this.result.set(null);
    this.coldStartNotice.set(false);
    this.goStatus.set({ state: 'checking', elapsed: 0 });
    this.nodeStatus.set({ state: 'checking', elapsed: 0 });
    const started = Date.now();
    const slowTimer = window.setTimeout(() => {
      if (this.busy()) this.coldStartNotice.set(true);
    }, 2500);
    const goTimer = window.setInterval(() => this.updateElapsed(this.goStatus, started), 500);
    const nodeTimer = window.setInterval(() => this.updateElapsed(this.nodeStatus, started), 500);

    try {
      await Promise.all([
        this.probe(goUrl, this.goStatus),
        this.probe(nodeUrl, this.nodeStatus),
      ]);
      this.phase.set('calculating');
      const expiresAt = sessionStorage.getItem('matrix-lab.token-expires-at');
      if (!this.accessToken || !expiresAt || Date.parse(expiresAt) <= Date.now()) { this.logout(); throw new Error('La sesión expiró. Inicia sesión nuevamente.'); }
      const response = await this.api.factorize(goUrl, matrix, this.accessToken);
      this.result.set(response);
    } catch (error) {
      const details = this.api.errorMessage(error);
      const sessionExpired = error instanceof Error && error.message.includes('sesión expiró');
      const unauthorized = (error as { status?: number })?.status === 401;
      if (sessionExpired || unauthorized) this.logout();
      this.requestError.set(sessionExpired ? 'La sesión expiró. Inicia sesión nuevamente.' : details.message);
      this.requestId.set(details.requestId ?? '');
    } finally {
      window.clearTimeout(slowTimer);
      window.clearInterval(goTimer);
      window.clearInterval(nodeTimer);
      this.busy.set(false);
      this.phase.set('idle');
      this.coldStartNotice.set(false);
    }
  }

  stateLabel(state: ServiceState): string {
    return ({ idle: 'En espera', checking: 'Conectando', ready: 'Disponible', slow: 'Iniciando', offline: 'Sin respuesta' })[state];
  }

  displayNumber(value: number): string {
    if (Object.is(value, -0) || Math.abs(value) < 1e-14) return '0';
    return new Intl.NumberFormat('es-MX', { maximumFractionDigits: 5 }).format(value);
  }

  private async probe(url: string, status: WritableSignal<ServiceStatus>): Promise<void> {
    try {
      await this.api.checkHealth(url);
      status.set({ state: 'ready', elapsed: status().elapsed });
    } catch {
      status.set({ state: 'offline', elapsed: status().elapsed });
    }
  }

  private updateElapsed(status: WritableSignal<ServiceStatus>, started: number): void {
    const current = status();
    if (current.state === 'checking') {
      const elapsed = Date.now() - started;
      status.set({ state: elapsed > 2500 ? 'slow' : 'checking', elapsed });
    }
  }

  private resizeMatrix(rows: number, columns: number): void {
    if (!Number.isInteger(rows) || !Number.isInteger(columns) || rows < 1 || columns < 1 || rows > 100 || columns > 100 || rows * columns > 10000) {
      this.inputError.set('El tamaño debe tener entre 1 y 100 filas y columnas, con un máximo de 10 000 celdas.');
      return;
    }
    const current = this.matrixCells();
    const resized = Array.from({ length: rows }, (_, row) =>
      Array.from({ length: columns }, (_, column) => current[row]?.[column] ?? ''),
    );
    this.rowCount.set(rows);
    this.columnCount.set(columns);
    this.matrixCells.set(resized);
    this.inputError.set('');
    this.requestError.set('');
    this.result.set(null);
  }

  private parseMatrix(cells: string[][]): number[][] {
    return cells.map((row, rowIndex) => row.map((cell, columnIndex) => {
      if (cell.trim() === '') throw new Error(`Completa la celda de la fila ${rowIndex + 1}, columna ${columnIndex + 1}.`);
      const value = Number(cell);
      if (!Number.isFinite(value)) throw new Error(`La celda de la fila ${rowIndex + 1}, columna ${columnIndex + 1} debe ser un número finito.`);
      return value;
    }));
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

  private readSetting(key: string, fallback: string): string {
    if (typeof localStorage === 'undefined') return fallback;
    return localStorage.getItem(key) ?? fallback;
  }
}
