import { TestBed } from '@angular/core/testing';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiService } from '../../shared/api.service';
import { FactorizationResponse } from '../../shared/models';
import { CalculationPageComponent } from './calculation-page.component';

describe('CalculationPageComponent', () => {
  let api: { checkHealth: ReturnType<typeof vi.fn>; factorize: ReturnType<typeof vi.fn>; errorMessage: ReturnType<typeof vi.fn> };

  beforeEach(async () => {
    api = {
      checkHealth: vi.fn().mockResolvedValue(undefined),
      factorize: vi.fn().mockResolvedValue({ shape: { rows: 3, columns: 2 }, q: [], r: [], statistics: {} } as unknown as FactorizationResponse),
      errorMessage: vi.fn().mockReturnValue({ message: 'API error' }),
    };
    await TestBed.configureTestingModule({
      imports: [CalculationPageComponent],
      providers: [{ provide: ApiService, useValue: api }],
    }).compileComponents();
  });

  it('renders the default matrix as editable cells', () => {
    const fixture = TestBed.createComponent(CalculationPageComponent);
    fixture.componentRef.setInput('accessToken', 'test-token');
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelectorAll('.matrix-input-table tbody input').length).toBe(6);
    expect(fixture.nativeElement.querySelector('textarea')).toBeNull();
  });

  it('resizes the grid while preserving cells that remain in range', () => {
    const fixture = TestBed.createComponent(CalculationPageComponent);
    fixture.componentRef.setInput('accessToken', 'test-token');
    const page = fixture.componentInstance;
    page.resizeRows(2);
    page.resizeColumns(3);
    fixture.detectChanges();
    expect(page.matrixCells()).toEqual([['1', '0', ''], ['0', '1', '']]);
    expect(fixture.nativeElement.querySelectorAll('.matrix-input-table tbody input').length).toBe(6);
  });

  it('does not submit when any matrix cell is empty', async () => {
    const fixture = TestBed.createComponent(CalculationPageComponent);
    fixture.componentRef.setInput('accessToken', 'test-token');
    await fixture.whenStable();
    fixture.componentInstance.setCellValue(0, 0, '');
    await fixture.componentInstance.factorize();
    expect(fixture.componentInstance.inputError()).toMatch(/celda.*fila 1, columna 1/i);
    expect(api.factorize).not.toHaveBeenCalled();
  });

  it('submits numeric cells and the access token to Go', async () => {
    const fixture = TestBed.createComponent(CalculationPageComponent);
    fixture.componentRef.setInput('accessToken', 'test-token');
    await fixture.whenStable();
    sessionStorage.setItem('matrix-lab.token-expires-at', new Date(Date.now() + 60_000).toISOString());
    await fixture.componentInstance.factorize();
    expect(api.factorize).toHaveBeenCalledWith('http://localhost:8080', [[1, 0], [0, 1], [1, 1]], 'test-token');
  });
});
