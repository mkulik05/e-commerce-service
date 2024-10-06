from locust import HttpUser, between, task
import json
import random

class WebsiteUser(HttpUser):
    wait_time = between(5, 15)

    def on_start(self):
       
        self.jwt = self.register_and_login_user()
        
    def register_and_login_user(self):

        login = f"tuser{self.environment.runner.user_count}-{random.randint(31, 9000000000)}"
        pwd = f"password123-{random.randint(31, 9000000000)}"

        register_url = "/auth/register"
        headers = {
            "Content-Type": "application/json"
        }
        register_data = {
            "login": login,
            "pwd": pwd
        }
        response = self.client.post(register_url, headers=headers, data=json.dumps(register_data))

        login_url = "/auth/login"
        login_data = {
            "login": login,
            "pwd": pwd
        }
        response = self.client.post(login_url, headers=headers, data=json.dumps(login_data))


        return response.json().get('token')

    @task
    def place_order(self):
        url = "/orders/order"
        headers = {
            "Content-Type": "application/json",
            "Authorization": self.jwt   for authorization
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
            "Authorization": self.jwt  
        }
        self.client.get(url, headers=headers)

    @task
    def get_items(self):
        self.client.get("/items/list?page=0")

    @task
    def get_items_list(self):
        self.client.get("/items/list?sort=price&sort_order=asc")