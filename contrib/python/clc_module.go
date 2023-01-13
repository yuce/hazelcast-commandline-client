package python

const PythonModule = `
import os
import logging
from pathlib import Path

import hazelcast
from hazelcast.discovery import HazelcastCloudDiscovery
from hazelcast.db import connect

import yaml

__all__ = ["client"]


def home():
    d = os.getenv("CLC_HOME")
    if d is not None:
        return d
    return Path.home().joinpath(".local/share/clc")


def config_path(name=None):
    if name is None:
        path = os.getenv("CLC_CONFIG")
        if path is None:
            raise Exception("Cannot deduce the configuration path, try setting CLC_CONFIG")
    else:
        path = home().joinpath("configs", f"{name}/config.yaml")
    return Path(path)


def load_config(path):
    try:
        with open(path) as f:
            cfg = yaml.load(f, Loader=yaml.Loader)
    except FileNotFoundError:
        cfg = None
    cfg = cfg or {}
    config_dir = path.parent
    ssl = cfg.get("ssl", {})
    if ssl:
        ssl["ca-path"] = str(config_dir.joinpath(config_dir, ssl["ca-path"]))
        ssl["cert-path"] = str(config_dir.joinpath(config_dir, ssl["cert-path"]))
        ssl["key-path"] = str(config_dir.joinpath(config_dir, ssl["key-path"]))
    return cfg


def make_client_config(cfg):
    cfg = cfg or {}
    ssl = cfg.get("ssl", {})
    cluster = cfg.get("cluster", {})
    d = dict(
        cluster_name=cluster.get("name", "dev"),
        cloud_discovery_token=cluster.get("discovery-token"),
        statistics_enabled=True,
        ssl_cafile=ssl.get("ca-path"),
        ssl_certfile=ssl.get("cert-path"),
        ssl_keyfile=ssl.get("key-path"),
        ssl_password=ssl.get("key-password"),
    )
    d["ssl_enabled"] = bool(d["ssl_certfile"])
    return d


#logging.basicConfig(level=logging.INFO)
HazelcastCloudDiscovery._CLOUD_URL_BASE = "api.viridian.hazelcast.com"

cfg = load_config(config_path())
client_cfg = make_client_config(cfg)
client = hazelcast.HazelcastClient(**client_cfg)
conn = connect(client_cfg)
`
