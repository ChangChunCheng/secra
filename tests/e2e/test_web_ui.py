import pytest
from playwright.sync_api import Page, expect

BASE_URL = "http://127.0.0.1:8081"

def test_home_page_dashboard(page: Page):
    page.goto(BASE_URL)
    # Check if logo is present and has the correct link
    logo = page.locator("header h1")
    expect(logo).to_contain_text("SECRA")
    
    # Check stats cards
    expect(page.locator("text=Total CVEs")).to_be_visible()
    expect(page.locator("text=Tracked Vendors")).to_be_visible()
    expect(page.locator("text=Monitored Products")).to_be_visible()

def test_user_registration_and_login(page: Page):
    import time
    username = f"testuser_{int(time.time())}"
    email = f"{username}@example.com"
    password = "testpassword"

    # Register
    page.goto(f"{BASE_URL}/register")
    page.fill('input[name="username"]', username)
    page.fill('input[name="email"]', email)
    page.fill('input[name="password"]', password)
    page.click('button:has-text("Register")')

    # Should redirect to login
    expect(page).to_have_url(f"{BASE_URL}/login")

    # Login
    page.fill('input[name="username"]', username)
    page.fill('input[name="password"]', password)
    page.click('button:has-text("Login")')

    # Should redirect to home and show profile
    expect(page).to_have_url(f"{BASE_URL}/")
    expect(page.locator(f"text=Profile ({username})")).to_be_visible()
    expect(page.locator("text=Logout")).to_be_visible()

def test_admin_access(page: Page):
    # Login as admin (assuming admin/adminpassword exists from previous setup)
    page.goto(f"{BASE_URL}/login")
    page.fill('input[name="username"]', "admin")
    page.fill('input[name="password"]', "adminpassword")
    page.click('button:has-text("Login")')

    # Check for admin-only link
    admin_link = page.locator('text=[Admin] Users')
    expect(admin_link).to_be_visible()
    
    # Navigate to admin users page
    admin_link.click()
    expect(page).to_have_url(f"{BASE_URL}/admin/users")
    expect(page.locator("text=User Administration")).to_be_visible()
    expect(page.get_by_role("cell", name="admin").first).to_be_visible()

def test_protected_routes_redirect(page: Page):
    # Try to access dashboard without login
    page.goto(f"{BASE_URL}/my/dashboard")
    expect(page).to_have_url(f"{BASE_URL}/login")

def test_cve_list_and_detail(page: Page):
    page.goto(f"{BASE_URL}/cves")
    expect(page.locator("text=CVE Intelligence Explorer")).to_be_visible()
    
    # Click on the first CVE link if any exist
    first_cve = page.locator("table tbody tr td a").first
    if first_cve.count() > 0:
        cve_id = first_cve.inner_text()
        first_cve.click()
        expect(page.locator(f"h2:has-text('{cve_id}')")).to_be_visible()
        expect(page.locator("text=Affected Products")).to_be_visible()
