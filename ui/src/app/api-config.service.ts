import { Injectable, InjectionToken, inject } from '@angular/core';

/**
 * API Configuration interface for dependency injection
 */
export interface ApiConfig {
  /** Base URL for HTTP API (e.g., 'http://localhost:9090') */
  apiHost?: string;
  /** API namespace (e.g., '/api/v1/mining') */
  apiNamespace?: string;
  /** Force HTTPS/WSS even on localhost */
  forceSecure?: boolean;
}

/**
 * Injection token for providing custom API configuration
 */
export const API_CONFIG = new InjectionToken<ApiConfig>('API_CONFIG');

/**
 * Service to provide consistent API URL configuration across the application.
 *
 * By default, it auto-detects the protocol and host from the current page location,
 * which allows the app to work correctly whether served over HTTP or HTTPS.
 *
 * The configuration can be customized by providing API_CONFIG in the app module:
 *
 * @example
 * // In app.config.ts
 * providers: [
 *   { provide: API_CONFIG, useValue: { apiHost: 'https://api.example.com' } }
 * ]
 */
@Injectable({
  providedIn: 'root'
})
export class ApiConfigService {
  private readonly config = inject(API_CONFIG, { optional: true });

  /** Default API namespace */
  private readonly defaultNamespace = '/api/v1/mining';

  /**
   * Get the base URL for HTTP API requests
   * @returns Full HTTP base URL (e.g., 'https://localhost:9090/api/v1/mining')
   */
  get apiBaseUrl(): string {
    const host = this.getApiHost();
    const namespace = this.config?.apiNamespace ?? this.defaultNamespace;
    return `${host}${namespace}`;
  }

  /**
   * Get the WebSocket URL for event streaming
   * @returns Full WebSocket URL (e.g., 'wss://localhost:9090/api/v1/mining/ws/events')
   */
  get wsUrl(): string {
    const host = this.getApiHost();
    const namespace = this.config?.apiNamespace ?? this.defaultNamespace;
    const protocol = host.startsWith('https') ? 'wss' : 'ws';
    // Replace http(s):// with ws(s)://
    const wsHost = host.replace(/^https?:\/\//, `${protocol}://`);
    return `${wsHost}${namespace}/ws/events`;
  }

  /**
   * Get the API host (protocol + hostname + port)
   */
  private getApiHost(): string {
    // If custom host is configured, use it
    if (this.config?.apiHost) {
      return this.config.apiHost;
    }

    // Auto-detect from current page location
    if (typeof window !== 'undefined' && window.location) {
      const { protocol, hostname, port } = window.location;

      // Determine if we should use secure protocol
      const isSecure = protocol === 'https:' || this.config?.forceSecure;
      const httpProtocol = isSecure ? 'https' : 'http';

      // Default to port 9090 if we're on a dev server (4200) or no port specified
      const apiPort = this.getApiPort(port);

      return `${httpProtocol}://${hostname}:${apiPort}`;
    }

    // Fallback for SSR or non-browser environments
    return 'http://localhost:9090';
  }

  /**
   * Determine the API port based on the current page port
   */
  private getApiPort(currentPort: string): string {
    // If we're on the Angular dev server (4200), use the API default port
    if (currentPort === '4200' || currentPort === '') {
      return '9090';
    }
    // Otherwise, assume we're served from the same port as the API
    return currentPort;
  }

  /**
   * Check if the connection is secure (HTTPS/WSS)
   */
  get isSecure(): boolean {
    if (this.config?.forceSecure) {
      return true;
    }
    if (typeof window !== 'undefined') {
      return window.location.protocol === 'https:';
    }
    return false;
  }
}
