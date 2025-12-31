import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { SniderMining } from './app';

describe('SniderMining', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SniderMining],
      providers: [provideHttpClient()]
    }).compileComponents();
  });

  it('should create the app', () => {
    const fixture = TestBed.createComponent(SniderMining);
    const app = fixture.componentInstance;
    expect(app).toBeTruthy();
  });
});
