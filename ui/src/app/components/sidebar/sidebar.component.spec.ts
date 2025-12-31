import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SidebarComponent } from './sidebar.component';

describe('SidebarComponent', () => {
  let component: SidebarComponent;
  let fixture: ComponentFixture<SidebarComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SidebarComponent]
    }).compileComponents();

    fixture = TestBed.createComponent(SidebarComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should start in expanded state', () => {
    expect(component.collapsed()).toBeFalse();
  });

  it('should start with mobile menu closed', () => {
    expect(component.mobileOpen()).toBeFalse();
  });

  it('should toggle collapsed state', () => {
    expect(component.collapsed()).toBeFalse();

    component.toggleCollapse();
    expect(component.collapsed()).toBeTrue();

    component.toggleCollapse();
    expect(component.collapsed()).toBeFalse();
  });

  it('should toggle mobile menu state', () => {
    expect(component.mobileOpen()).toBeFalse();

    component.toggleMobileMenu();
    expect(component.mobileOpen()).toBeTrue();

    component.toggleMobileMenu();
    expect(component.mobileOpen()).toBeFalse();
  });

  it('should close mobile menu', () => {
    component.mobileOpen.set(true);
    expect(component.mobileOpen()).toBeTrue();

    component.closeMobileMenu();
    expect(component.mobileOpen()).toBeFalse();
  });

  it('should emit route change on navigate', () => {
    const emitSpy = spyOn(component.routeChange, 'emit');

    component.navigate('workers');

    expect(emitSpy).toHaveBeenCalledWith('workers');
  });

  it('should emit route change and close mobile menu on navigateAndClose', () => {
    const emitSpy = spyOn(component.routeChange, 'emit');
    component.mobileOpen.set(true);

    component.navigateAndClose('profiles');

    expect(emitSpy).toHaveBeenCalledWith('profiles');
    expect(component.mobileOpen()).toBeFalse();
  });

  it('should have correct number of nav items', () => {
    expect(component.navItems.length).toBe(7);
  });

  it('should have required routes', () => {
    const routes = component.navItems.map(item => item.route);

    expect(routes).toContain('dashboard');
    expect(routes).toContain('workers');
    expect(routes).toContain('console');
    expect(routes).toContain('pools');
    expect(routes).toContain('profiles');
    expect(routes).toContain('miners');
    expect(routes).toContain('nodes');
  });

  it('should render navigation items', () => {
    const navItems = fixture.nativeElement.querySelectorAll('.nav-item');
    expect(navItems.length).toBe(7);
  });

  it('should apply active class to current route', () => {
    fixture.componentRef.setInput('currentRoute', 'workers');
    fixture.detectChanges();

    const activeItem = fixture.nativeElement.querySelector('.nav-item.active');
    expect(activeItem).toBeTruthy();
  });

  it('should show logo text when expanded', () => {
    component.collapsed.set(false);
    fixture.detectChanges();

    const logoText = fixture.nativeElement.querySelector('.logo-text');
    expect(logoText).toBeTruthy();
    expect(logoText.textContent).toContain('Mining');
  });

  it('should hide nav labels when collapsed on desktop', () => {
    component.collapsed.set(true);
    component.mobileOpen.set(false);
    fixture.detectChanges();

    const navLabels = fixture.nativeElement.querySelectorAll('.nav-label');
    expect(navLabels.length).toBe(0);
  });
});
