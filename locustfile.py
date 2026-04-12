from locust import HttpUser, task, between
class WebUser(HttpUser):
    wait_time = between(0.05,
                        0.2)
    @task(5)
    def index(self): self.client.get("/", headers={"Connection":"close"})
    @task(2)
    def page2(self): self.client.get("/page2.html", headers={"Connection":"close"})
    @task(1)
    def nope(self):  self.client.get("/nope.html", headers={"Connection":"close"})
