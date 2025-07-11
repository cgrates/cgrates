### CGRateS Monitoring Setup Playbook

#### Inventory (inventory.ini):

```ini
[monit]
foo.example.com ansible_host=myserverip ansible_port=22 ansible_user=myuser

[monit:vars]
install_or_update_cgrates=true
```

#### Run playbook:

```bash
ansible-playbook -i inventory.ini /path/to/playbook/main.yml
```

#### Access Grafana:

```bash
ssh -L 8080:localhost:3000 myuser@myserverip
```

Browse to http://localhost:8080 and login

username: admin
password: admin

#### Components and their ports:
CGRateS: 2012
Node Exporter: 9100
Prometheus: 9090
Grafana: 3000

#### Imported Grafana dashboards:
- Go Metrics (ID: 13240) - a custom solution for this one would be preferred
- Node Exporter (ID: 1860)

> [!NOTE]
> Go Metrics tracks both the node_exporter and prometheus alongside CGRateS (all written in go). Make sure that job "cgrates" is the one selected.

Services can be managed via systemd
