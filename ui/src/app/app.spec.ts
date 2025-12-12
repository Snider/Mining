import { TestBed } from '@angular/core/testing';
import { SniderMining } from './app';
import { MinerService } from './miner.service';
import { signal } from '@angular/core';

describe('SniderMining', () => {
  beforeEach(async () => {
    const minerServiceMock = {
      state: signal({
          needsSetup: false,
          apiAvailable: true,
          systemInfo: {},
          manageableMiners: [],
          installedMiners: [],
          runningMiners: [],
          profiles: []
      }),
      forceRefreshState: jasmine.createSpy('forceRefreshState')
    };

    await TestBed.configureTestingModule({
      imports: [SniderMining],
      providers: [
        { provide: MinerService, useValue: minerServiceMock }
      ]
    }).compileComponents();
  });

  it('should create the app', () => {
    const fixture = TestBed.createComponent(SniderMining);
    const app = fixture.componentInstance;
    expect(app).toBeTruthy();
  });
});
