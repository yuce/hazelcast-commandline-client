package notebook

const PythonModule = `
import os
import logging
from pathlib import Path

import hazelcast
from hazelcast.discovery import HazelcastCloudDiscovery
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
    with open(path) as f:
        cfg = yaml.load(f, Loader=yaml.Loader)
    config_dir = path.parent
    ssl = cfg["ssl"]
    ssl["ca-path"] = str(config_dir.joinpath(config_dir, ssl["ca-path"]))
    ssl["cert-path"] = str(config_dir.joinpath(config_dir, ssl["cert-path"]))
    ssl["key-path"] = str(config_dir.joinpath(config_dir, ssl["key-path"]))
    return cfg


def make_client_config(cfg):
    return dict(
        cluster_name=cfg["cluster"]["name"],
        cloud_discovery_token=cfg["cluster"]["viridian-token"],
        statistics_enabled=True,
        ssl_enabled=True,
        ssl_cafile=cfg["ssl"]["ca-path"],
        ssl_certfile=cfg["ssl"]["cert-path"],
        ssl_keyfile=cfg["ssl"]["key-path"],
        ssl_password="12ee6ff601a",
    )


logging.basicConfig(level=logging.INFO)
HazelcastCloudDiscovery._CLOUD_URL_BASE = "api.viridian.hazelcast.com"

cfg = load_config(config_path())
client_cfg = make_client_config(cfg)
client = hazelcast.HazelcastClient(**client_cfg)

`
