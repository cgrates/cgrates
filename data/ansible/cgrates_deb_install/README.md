# CGRates Package Installation

Installs CGRates 1.0 from deb package and configures from private GitHub repo.

## Setup

1. Configure `inventory.ini` with your server and GitHub details
2. Run: `ansible-playbook -i inventory.ini main.yaml`

## Repository Structure

```
config-repo/
├── node1/
│   ├── etc/
│   │   └── cgrates/
│   │       └── cgrates.json
│   └── tp/
└── node2/
    ├── etc/
    │   └── cgrates/
    └── tp/
```

The playbook clones the repo to `/opt/{repo-name}` and creates symlinks:
- `/etc/cgrates` → `/opt/{repo-name}/node1/etc/cgrates`
