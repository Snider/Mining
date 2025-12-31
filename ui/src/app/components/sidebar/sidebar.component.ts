import { Component, signal, output, input, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';
import { IconComponent, IconName } from './icon.component';

interface NavItem {
  id: string;
  label: string;
  icon: IconName;
  route: string;
}

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, IconComponent],
  template: `
    <!-- Mobile menu button (visible on small screens) -->
    <button class="mobile-menu-btn" (click)="toggleMobileMenu()">
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
      </svg>
    </button>

    <!-- Mobile backdrop -->
    @if (mobileOpen()) {
      <div class="mobile-backdrop" (click)="closeMobileMenu()"></div>
    }

    <aside class="sidebar" [class.collapsed]="collapsed()" [class.mobile-open]="mobileOpen()">
      <!-- Logo / Brand -->
      <div class="sidebar-header">
        <div class="logo">
          <svg class="w-8 h-8 text-accent-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
          </svg>
          @if (!collapsed() || mobileOpen()) {
            <span class="logo-text">Mining</span>
          }
        </div>
        <button class="collapse-btn desktop-only" (click)="toggleCollapse()">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            @if (collapsed()) {
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7"/>
            } @else {
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7"/>
            }
          </svg>
        </button>
        <button class="collapse-btn mobile-only" (click)="closeMobileMenu()">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <!-- Navigation -->
      <nav class="sidebar-nav">
        @for (item of navItems; track item.id) {
          <button
            class="nav-item"
            [class.active]="currentRoute() === item.route"
            (click)="navigateAndClose(item.route)"
            [title]="collapsed() && !mobileOpen() ? item.label : ''">
            <span class="nav-icon">
              <app-icon [name]="item.icon" />
            </span>
            @if (!collapsed() || mobileOpen()) {
              <span class="nav-label">{{ item.label }}</span>
            }
          </button>
        }
      </nav>

      <!-- Footer with miner switcher placeholder -->
      <div class="sidebar-footer">
        @if (!collapsed() || mobileOpen()) {
          <div class="miner-status">
            <div class="status-indicator online"></div>
            <span class="status-text">Mining Active</span>
          </div>
        } @else {
          <div class="status-indicator online mx-auto"></div>
        }
      </div>
    </aside>
  `,
  styles: [`
    /* Mobile menu button */
    .mobile-menu-btn {
      display: none;
      position: fixed;
      top: 0.75rem;
      left: 0.75rem;
      z-index: 1001;
      padding: 0.5rem;
      background: var(--color-surface-200);
      border: 1px solid rgb(37 37 66 / 0.3);
      border-radius: 0.375rem;
      color: white;
      cursor: pointer;
    }

    .mobile-backdrop {
      display: none;
      position: fixed;
      inset: 0;
      background: rgba(0, 0, 0, 0.5);
      z-index: 999;
    }

    .mobile-only {
      display: none;
    }

    .sidebar {
      display: flex;
      flex-direction: column;
      width: var(--spacing-sidebar-expanded, 200px);
      height: 100vh;
      background: var(--color-surface-200);
      border-right: 1px solid rgb(37 37 66 / 0.2);
      transition: width 0.2s ease, transform 0.3s ease;
    }

    .sidebar.collapsed {
      width: var(--spacing-sidebar, 56px);
    }

    .sidebar-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 1rem;
      border-bottom: 1px solid rgb(37 37 66 / 0.2);
    }

    .logo {
      display: flex;
      align-items: center;
      gap: 0.75rem;
    }

    .logo-text {
      font-size: 1.125rem;
      font-weight: 600;
      color: white;
    }

    .collapse-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: transparent;
      border: none;
      border-radius: 0.375rem;
      color: #94a3b8;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .collapse-btn:hover {
      background: rgb(37 37 66 / 0.5);
      color: white;
    }

    .collapsed .collapse-btn {
      margin: 0 auto;
    }

    .sidebar-nav {
      flex: 1;
      display: flex;
      flex-direction: column;
      padding: 0.5rem;
      gap: 0.25rem;
      overflow-y: auto;
    }

    .nav-item {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      padding: 0.625rem 0.75rem;
      border-radius: 0.5rem;
      border: none;
      background: transparent;
      color: #94a3b8;
      cursor: pointer;
      transition: all 0.15s ease;
      width: 100%;
      text-align: left;
    }

    .nav-item:hover {
      color: white;
      background: rgb(37 37 66 / 0.5);
    }

    .nav-item.active {
      background: rgb(0 212 255 / 0.1);
      color: var(--color-accent-400);
      border-left: 2px solid var(--color-accent-500);
    }

    .collapsed .nav-item {
      justify-content: center;
      padding: 0.625rem;
    }

    .nav-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 20px;
      height: 20px;
      flex-shrink: 0;
    }

    .nav-label {
      font-size: 0.875rem;
      font-weight: 500;
    }

    .sidebar-footer {
      padding: 1rem;
      border-top: 1px solid rgb(37 37 66 / 0.2);
    }

    .miner-status {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .status-indicator {
      width: 8px;
      height: 8px;
      border-radius: 9999px;
    }

    .status-indicator.online {
      background: var(--color-success-500);
      box-shadow: 0 0 8px var(--color-success-500);
    }

    .status-text {
      font-size: 0.75rem;
      color: #94a3b8;
    }

    /* Mobile styles */
    @media (max-width: 768px) {
      .mobile-menu-btn {
        display: flex;
      }

      .mobile-backdrop {
        display: block;
      }

      .desktop-only {
        display: none;
      }

      .mobile-only {
        display: flex;
      }

      .sidebar {
        position: fixed;
        left: 0;
        top: 0;
        z-index: 1000;
        width: 280px;
        transform: translateX(-100%);
      }

      .sidebar.mobile-open {
        transform: translateX(0);
      }

      .sidebar.collapsed {
        width: 280px;
      }

      .collapsed .nav-item {
        justify-content: flex-start;
        padding: 0.625rem 0.75rem;
      }
    }
  `]
})
export class SidebarComponent {
  collapsed = signal(false);
  mobileOpen = signal(false);
  currentRoute = input<string>('dashboard');
  routeChange = output<string>();

  @HostListener('window:resize')
  onResize() {
    // Close mobile menu on resize to larger screens
    if (window.innerWidth > 768 && this.mobileOpen()) {
      this.mobileOpen.set(false);
    }
  }

  navItems: NavItem[] = [
    { id: 'dashboard', label: 'Dashboard', route: 'dashboard', icon: 'dashboard' },
    { id: 'workers', label: 'Workers', route: 'workers', icon: 'workers' },
    { id: 'console', label: 'Console', route: 'console', icon: 'console' },
    { id: 'pools', label: 'Pools', route: 'pools', icon: 'pools' },
    { id: 'profiles', label: 'Profiles', route: 'profiles', icon: 'profiles' },
    { id: 'miners', label: 'Miners', route: 'miners', icon: 'miners' },
    { id: 'nodes', label: 'Nodes', route: 'nodes', icon: 'nodes' }
  ];

  toggleCollapse() {
    this.collapsed.update(v => !v);
  }

  toggleMobileMenu() {
    this.mobileOpen.update(v => !v);
  }

  closeMobileMenu() {
    this.mobileOpen.set(false);
  }

  navigate(route: string) {
    this.routeChange.emit(route);
  }

  navigateAndClose(route: string) {
    this.routeChange.emit(route);
    this.closeMobileMenu();
  }
}
