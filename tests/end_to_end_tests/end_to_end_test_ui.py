from selenium import webdriver
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from webdriver_manager.chrome import ChromeDriverManager



GUI_URL = "http://localhost:15000/public"

REGISTER = "http://127.0.0.1:15000/register"

gecko_path = "C:\\Users\\roman\\Downloads\\geckodriver-v0.32.0-win32\\geckodriver.exe"


firefox_options = Options()
firefox_options.add_argument("--headless")

driver = webdriver.Chrome(service=Service(ChromeDriverManager().install()), options=firefox_options)
driver.get(REGISTER)

def test_register_user_via_gui():

    search_bar = driver.find_element_by_name("username")
    search_bar.clear()
    search_bar.send_keys("SelTester")

    search_bar = driver.find_element_by_name("email")
    search_bar.clear()
    search_bar.send_keys("SelTester@test.com")

    search_bar = driver.find_element_by_name("password")
    search_bar.clear()
    search_bar.send_keys("test")

    search_bar = driver.find_element_by_name("password2")
    search_bar.clear()
    search_bar.send_keys("test")

    search_bar.send_keys(Keys.RETURN)

    print(driver.current_url)

    driver.close()