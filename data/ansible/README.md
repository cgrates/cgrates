# CGRateS Ansible roles

On your Ansible control machine:

```sh
ansible-galaxy collection install 'git+https://github.com/cgrates/cgrates.git#/data/ansible/,main'
```

By default, `ansible-galaxy` installs the collection in `~/.ansible/collections`
on the control machine. Ansible finds it from any directory without
`--playbook-dir`.

Run an installed role:

```text
ansible all -i HOST, --playbook-dir . -u USER -m ansible.builtin.import_role -a 'name=cgrates.deployment.ROLE' [-K] [-e 'VARIABLE=VALUE']
```

Replace `HOST`, `USER` and `ROLE` with your host, SSH user and role name.
The user needs sudo access; add `-K` if it requires a password.
Use `-e` to override variables.

Examples for Debian 13 amd64 (replace `server` and `admin`):

Install CGRateS using a playbook. Save this as `cgrates.yml`:

```yaml
- hosts: all
  roles:
    - cgrates.deployment.cgrates
```

```sh
ansible-playbook -i server, -u admin cgrates.yml
```

Install Node Exporter with a direct command. Its service stays stopped by
default:

```sh
ansible all -i server, -u admin -m ansible.builtin.import_role -a 'name=cgrates.deployment.node_exporter'
```

CGRateS installs from source by default.
The Go role checks `/usr/local/go/bin/go`. If its version matches `go_version`,
it is reused. Otherwise the role installs the requested version.
Set `install_go=false` to skip installing Go.
CGRateS source builds and Go cache cleanup need environment facts. Playbooks
gather those facts by default. The `ansible` command does not gather them
automatically.

Databases are not installed by default. `cgrates_dbs` selects setup scripts that
can delete existing data; it does not install database packages.
See each role's defaults and tasks in [`roles/`](roles/).
