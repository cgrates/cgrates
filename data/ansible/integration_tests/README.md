# Ansible for CGRateS integration tests 

Steps for running this ansible:
1. Configure a debian based virtual machine so you can connect to it without needing the password
2. Edit your ansible host file with the IP and the user of the machine like so:
```
[all]
local ansible_host=192.168.56.203 ansible_ssh_user=trial97

[all:vars]
user=trial97
```

3. Run the ansible:
```
ansible-playbook main.yaml
```

4. Done!
