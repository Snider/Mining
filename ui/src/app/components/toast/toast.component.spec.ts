import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ToastComponent } from './toast.component';
import { NotificationService } from '../../notification.service';

describe('ToastComponent', () => {
  let component: ToastComponent;
  let fixture: ComponentFixture<ToastComponent>;
  let notificationService: NotificationService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ToastComponent],
      providers: [NotificationService]
    }).compileComponents();

    fixture = TestBed.createComponent(ToastComponent);
    component = fixture.componentInstance;
    notificationService = TestBed.inject(NotificationService);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should display no toasts when there are no notifications', () => {
    const toasts = fixture.nativeElement.querySelectorAll('.toast');
    expect(toasts.length).toBe(0);
  });

  it('should display a success toast', () => {
    notificationService.success('Success message', 'Success');
    fixture.detectChanges();

    const toasts = fixture.nativeElement.querySelectorAll('.toast');
    expect(toasts.length).toBe(1);

    const toast = toasts[0];
    expect(toast.classList.contains('toast-success')).toBeTrue();
    expect(toast.textContent).toContain('Success message');
    expect(toast.textContent).toContain('Success');
  });

  it('should display an error toast', () => {
    notificationService.error('Error message', 'Error');
    fixture.detectChanges();

    const toast = fixture.nativeElement.querySelector('.toast-error');
    expect(toast).toBeTruthy();
    expect(toast.textContent).toContain('Error message');
  });

  it('should display a warning toast', () => {
    notificationService.warning('Warning message', 'Warning');
    fixture.detectChanges();

    const toast = fixture.nativeElement.querySelector('.toast-warning');
    expect(toast).toBeTruthy();
    expect(toast.textContent).toContain('Warning message');
  });

  it('should display an info toast', () => {
    notificationService.info('Info message', 'Info');
    fixture.detectChanges();

    const toast = fixture.nativeElement.querySelector('.toast-info');
    expect(toast).toBeTruthy();
    expect(toast.textContent).toContain('Info message');
  });

  it('should display multiple toasts', () => {
    notificationService.success('Message 1');
    notificationService.error('Message 2');
    notificationService.warning('Message 3');
    fixture.detectChanges();

    const toasts = fixture.nativeElement.querySelectorAll('.toast');
    expect(toasts.length).toBe(3);
  });

  it('should dismiss a toast when close button is clicked', () => {
    notificationService.success('Test message');
    fixture.detectChanges();

    let toasts = fixture.nativeElement.querySelectorAll('.toast');
    expect(toasts.length).toBe(1);

    const closeButton = fixture.nativeElement.querySelector('.toast-close');
    closeButton.click();
    fixture.detectChanges();

    toasts = fixture.nativeElement.querySelectorAll('.toast');
    expect(toasts.length).toBe(0);
  });

  it('should have correct icon for each notification type', () => {
    notificationService.success('Success');
    notificationService.error('Error');
    fixture.detectChanges();

    const toasts = fixture.nativeElement.querySelectorAll('.toast');
    toasts.forEach((toast: Element) => {
      const icon = toast.querySelector('.toast-icon svg');
      expect(icon).toBeTruthy();
    });
  });
});
