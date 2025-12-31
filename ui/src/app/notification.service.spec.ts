import { TestBed, fakeAsync, tick } from '@angular/core/testing';
import { NotificationService } from './notification.service';

describe('NotificationService', () => {
  let service: NotificationService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [NotificationService]
    });
    service = TestBed.inject(NotificationService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should start with no notifications', () => {
    expect(service.notifications().length).toBe(0);
  });

  it('should add a success notification', () => {
    service.success('Test message', 'Test title');

    const notifications = service.notifications();
    expect(notifications.length).toBe(1);
    expect(notifications[0].type).toBe('success');
    expect(notifications[0].message).toBe('Test message');
    expect(notifications[0].title).toBe('Test title');
  });

  it('should add an error notification', () => {
    service.error('Error message', 'Error title');

    const notifications = service.notifications();
    expect(notifications.length).toBe(1);
    expect(notifications[0].type).toBe('error');
    expect(notifications[0].message).toBe('Error message');
  });

  it('should add a warning notification', () => {
    service.warning('Warning message');

    const notifications = service.notifications();
    expect(notifications.length).toBe(1);
    expect(notifications[0].type).toBe('warning');
  });

  it('should add an info notification', () => {
    service.info('Info message');

    const notifications = service.notifications();
    expect(notifications.length).toBe(1);
    expect(notifications[0].type).toBe('info');
  });

  it('should assign unique IDs to notifications', () => {
    service.success('Message 1');
    service.success('Message 2');
    service.success('Message 3');

    const notifications = service.notifications();
    const ids = notifications.map(n => n.id);
    const uniqueIds = new Set(ids);

    expect(uniqueIds.size).toBe(3);
  });

  it('should dismiss a notification by ID', () => {
    service.success('Message 1');
    service.success('Message 2');

    const firstId = service.notifications()[0].id;
    service.dismiss(firstId);

    const remaining = service.notifications();
    expect(remaining.length).toBe(1);
    expect(remaining[0].id).not.toBe(firstId);
  });

  it('should dismiss all notifications', () => {
    service.success('Message 1');
    service.error('Message 2');
    service.warning('Message 3');

    expect(service.notifications().length).toBe(3);

    service.dismissAll();

    expect(service.notifications().length).toBe(0);
  });

  it('should auto-dismiss success notifications after duration', fakeAsync(() => {
    service.success('Test message', undefined, 1000);

    expect(service.notifications().length).toBe(1);

    tick(1000);

    expect(service.notifications().length).toBe(0);
  }));

  it('should auto-dismiss error notifications after longer duration', fakeAsync(() => {
    service.error('Error message', undefined, 2000);

    expect(service.notifications().length).toBe(1);

    tick(1999);
    expect(service.notifications().length).toBe(1);

    tick(1);
    expect(service.notifications().length).toBe(0);
  }));

  it('should handle multiple notifications with different durations', fakeAsync(() => {
    service.success('Quick', undefined, 500);
    service.error('Slow', undefined, 1500);

    expect(service.notifications().length).toBe(2);

    tick(500);
    expect(service.notifications().length).toBe(1);
    expect(service.notifications()[0].message).toBe('Slow');

    tick(1000);
    expect(service.notifications().length).toBe(0);
  }));
});
