from locust import HttpUser, between, task
import json
import random

class WebsiteUser(HttpUser):
    wait_time = between(5, 15)

    def on_start(self):
        # Register a new user and log in to obtain JWT
        self.jwt = self.register_and_login_user()
        
    def register_and_login_user(self):
        # Generate a unique login
        login = f"tuser{self.user_id()}-{random.randint(31, 9000000000)}"
        pwd = "password123"

        # Register the user
        register_url = "/auth/register"
        headers = {
            "Content-Type": "application/json"
        }
        register_data = {
            "login": login,
            "pwd": pwd
        }
        response = self.client.post(register_url, headers=headers, data=json.dumps(register_data))

        # Log in the user to get JWT
        login_url = "/auth/login"
        login_data = {
            "login": login,
            "pwd": pwd
        }
        response = self.client.post(login_url, headers=headers, data=json.dumps(login_data))

        # Assuming the JWT is returned in the response JSON under the key 'token'
        return response.json().get('token')

    @task
    def place_order(self):
        url = "/orders/order"
        headers = {
            "Content-Type": "application/json",
            "Authorization": self.jwt  # Use the JWT for authorization
        }
        data = {
            "items": {
                "1": 2,
                "2": 3,
                "4": 10
            },
            "delivery_addr": "123 Main Sfdnsiofnmdskt"
        }
        
        self.client.post(url, headers=headers, data=json.dumps(data))

    @task
    def get_orders(self):
        url = "/orders/list"
        headers = {
            "Authorization": self.jwt  # Use the JWT for authorization
        }
        self.client.get(url, headers=headers)

    @task
    def get_items(self):
        self.client.get("/items/list?page=0")

    @task
    def get_items_list(self):
        self.client.get("/items/list?sort=price&sort_order=asc")

    def user_id(self):
        # A simple way to generate a unique user ID based on the instance ID
        return self.environment.runner.user_count