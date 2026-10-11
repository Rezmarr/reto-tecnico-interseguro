import { TestBed } from '@angular/core/testing';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiService } from '../../shared/api.service';
import { FactorizationResponse } from '../../shared/models';
import { CalculationPageComponent } from './calculation-page.component';

describe('Componente de cálculo', () => {
  let api: { checkHealth: ReturnType<typeof vi.fn>; factorize: ReturnType<typeof vi.fn>; errorMessage: ReturnType<typeof vi.fn> };

  beforeEach(async () => {
    api = {
      checkHealth: vi.fn().mockResolvedValue(undefined),
      factorize: vi.fn().mockResolvedValue({ shape: { rows: 3, columns: 2 }, q: [], r: [], statistics: {} } as unknown as FactorizationResponse),
      errorMessage: vi.fn().mockReturnValue({ message: 'Error de API' }),
    };
    await TestBed.configureTestingModule({
      imports: [CalculationPageComponent],
      providers: [{ provide: ApiService, useValue: api }],
    }).compileComponents();
  });

  it('muestra la matriz inicial en celdas editables', () => {
    const fixture = TestBed.createComponent(CalculationPageComponent);
    fixture.componentRef.setInput('accessToken', 'test-token');
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelectorAll('.matrix-input-table tbody input').length).toBe(6);
    expect(fixture.nativeElement.querySelector('textarea')).toBeNull();
  });

  it('redimensiona la cuadrícula y conserva las celdas que siguen dentro del rango', () => {
    const fixture = TestBed.createComponent(CalculationPageComponent);
    fixture.componentRef.setInput('accessToken', 'test-token');
    const page = fixture.componentInstance;
    page.resizeRows(2);
    page.resizeColumns(3);
    fixture.detectChanges();
    expect(page.matrixCells()).toEqual([['1', '0', ''], ['0', '1', '']]);
    expect(fixture.nativeElement.querySelectorAll('.matrix-input-table tbody input').length).toBe(6);
  });

  it('no envía la matriz si alguna celda está vacía', async () => {
    const fixture = TestBed.createComponent(CalculationPageComponent);
    fixture.componentRef.setInput('accessToken', 'test-token');
    await fixture.whenStable();
    fixture.componentInstance.setCellValue(0, 0, '');
    await fixture.componentInstance.factorize();
    expect(fixture.componentInstance.inputError()).toMatch(/celda.*fila 1, columna 1/i);
    expect(api.factorize).not.toHaveBeenCalled();
  });

  it('envía las celdas numéricas y el token de acceso a Go', async () => {
    const fixture = TestBed.createComponent(CalculationPageComponent);
    fixture.componentRef.setInput('accessToken', 'test-token');
    await fixture.whenStable();
    sessionStorage.setItem('matrix-lab.token-expires-at', new Date(Date.now() + 60_000).toISOString());
    await fixture.componentInstance.factorize();
    expect(api.factorize).toHaveBeenCalledWith('http://localhost:8080', [[1, 0], [0, 1], [1, 1]], 'test-token');
  });
});
