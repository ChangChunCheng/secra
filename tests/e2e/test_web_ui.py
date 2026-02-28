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

def test_vendor_and_product_lists(page: Page):
    # Check Vendors
    page.goto(f"{BASE_URL}/vendors")
    expect(page.locator("text=Tracked Vendors")).to_be_visible()
    
    # Check Products
    page.goto(f"{BASE_URL}/products")
    expect(page.locator("text=Monitored Products")).to_be_visible()

def test_pagination(page: Page):
    page.goto(f"{BASE_URL}/cves")
    expect(page.get_by_role("heading", name="CVE Intelligence Feed (Page 1)")).to_be_visible()
    
    next_btn = page.locator('text=Next →')
    if next_btn.is_visible() and next_btn.is_enabled():
        next_btn.click()
        expect(page.get_by_role("heading", name="CVE Intelligence Feed (Page 2)")).to_be_visible()

def test_protected_routes_redirect(page: Page):
    # Try to access dashboard without login
    page.goto(f"{BASE_URL}/my/dashboard")
    expect(page).to_have_url(f"{BASE_URL}/login")

def test_cve_creation(page: Page):
    # Login first
    page.goto(f"{BASE_URL}/login")
    page.fill('input[name="username"]', "admin")
    page.fill('input[name="password"]', "adminpassword")
    page.click('button:has-text("Login")')

    # Go to new CVE page
    page.goto(f"{BASE_URL}/cves/new")
    import time
    # Timestamp is 10 digits, need to keep total under 16. Prefix 'T-' is 2 chars.
    cve_id = f"T-{int(time.time())}"
    page.fill('input[name="source_uid"]', cve_id)
    page.fill('input[name="title"]', "E2E Test Vulnerability")
    page.fill('textarea[name="description"]', "This is a vulnerability created by E2E test.")
    page.select_option('select[name="severity"]', "CRITICAL")
    page.fill('input[name="cvss_score"]', "9.8")
    page.click('button:has-text("Publish CVE")')

    # Should redirect to detail page
    expect(page.locator(f"h2:has-text('{cve_id}')")).to_be_visible()
    expect(page.get_by_text("CRITICAL", exact=True)).to_be_visible()
    expect(page.get_by_text("9.8", exact=True)).to_be_visible()

def test_product_subscription(page: Page):
    # Login
    page.goto(f"{BASE_URL}/login")
    page.fill('input[name="username"]', "admin")
    page.fill('input[name="password"]', "adminpassword")
    page.click('button:has-text("Login")')

    # Go to a CVE with products
    page.goto(f"{BASE_URL}/cves")
    # Find first CVE link and go to detail
    first_cve = page.locator("table tbody tr td a").first
    first_cve.click()

    # Check if there's a subscribe button
    subscribe_btn = page.locator('button:has-text("Subscribe")').first
    if subscribe_btn.count() > 0:
        subscribe_btn.click()
        # Should stay on page or redirect back
        # Go to My Dashboard to verify
        page.goto(f"{BASE_URL}/my/dashboard")
        expect(page.locator("text=Active Subscriptions")).to_be_visible()
        expect(page.locator("table tbody tr")).to_have_count(1, timeout=10000)
