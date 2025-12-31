import { Injectable, signal, computed } from '@angular/core';

export type NotificationType = 'success' | 'error' | 'warning' | 'info';

export interface Notification {
  id: number;
  type: NotificationType;
  message: string;
  title?: string;
  duration?: number;
}

@Injectable({
  providedIn: 'root'
})
export class NotificationService {
  private notificationId = 0;
  private _notifications = signal<Notification[]>([]);

  public notifications = computed(() => this._notifications());

  /**
   * Show a success notification
   */
  success(message: string, title?: string, duration = 4000) {
    this.show({ type: 'success', message, title, duration });
  }

  /**
   * Show an error notification
   */
  error(message: string, title?: string, duration = 6000) {
    this.show({ type: 'error', message, title, duration });
  }

  /**
   * Show a warning notification
   */
  warning(message: string, title?: string, duration = 5000) {
    this.show({ type: 'warning', message, title, duration });
  }

  /**
   * Show an info notification
   */
  info(message: string, title?: string, duration = 4000) {
    this.show({ type: 'info', message, title, duration });
  }

  /**
   * Show a notification
   */
  private show(notification: Omit<Notification, 'id'>) {
    const id = ++this.notificationId;
    const newNotification: Notification = { ...notification, id };

    this._notifications.update(notifications => [...notifications, newNotification]);

    // Auto-dismiss after duration
    if (notification.duration && notification.duration > 0) {
      setTimeout(() => this.dismiss(id), notification.duration);
    }
  }

  /**
   * Dismiss a notification by ID
   */
  dismiss(id: number) {
    this._notifications.update(notifications =>
      notifications.filter(n => n.id !== id)
    );
  }

  /**
   * Dismiss all notifications
   */
  dismissAll() {
    this._notifications.set([]);
  }
}
