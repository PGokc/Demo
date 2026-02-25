#!/usr/bin/env python3
# encoding=utf-8
from __future__ import print_function  # Python2 print compatibility

import argparse
import json
import requests
import time
import sys
import logging

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)

# 创建一个 StreamHandler，用于输出到控制台
console_handler = logging.StreamHandler()
console_handler.setLevel(logging.DEBUG)

# 将 StreamHandler 添加到 logger 中
logger.addHandler(console_handler)


def output_param(key, value):
    s = "JOB_RETURN::{key}={value}::JOB_RETURN".format(key=key, value=value)
    logger.info(s)


class RdsOperation(object):
    base_url = {
        # 'prod': "http://rdsmgr.default.svc.storage-cp.org:8089",
        'prod': "http://10.31.135.91:30889",
        'test': "http://rdsmgr2.default.svc.storage-cp.org:8089"
    }

    def __init__(self, env='prod'):
        self.url = self.base_url[env]

    def get_operation_by_name(self, name):
        url = "{0}/api/v1/operation/query/{1}".format(self.url, name)
        try:
            res = requests.get(url)
            if res.status_code != 200:
                raise Exception(
                    "get operation by name failed, res: {0}".format(res.text))
            return res.json().get("data")
        except Exception as e:
            raise e

    def create_operation(self, type, data):
        url = "{0}/api/v1/operation/split/create/{1}".format(self.url, type)
        try:
            res = requests.post(url, json=data)
            if res.status_code != 200:
                raise Exception("create operation failed, res: {0}".format(
                    res.text))
            output_param("log_id", res.headers.get("x-tt-logid"))
            return res.json().get("data")
        except Exception as e:
            raise e

    def create_operation_and_print_result(self, type, data):
        try:
            res = self.create_operation(type, data)
            name = res.get("name")
            output_param("operation_name", name)
            return name
        except Exception as e:
            raise e

    def print_output_params(self, data):
        for key, value in data.items():
            if value != "":
                output_param(key, value)
            if ',' in value:
                for v in value.split(','):
                    logger.info("{} item:  {}".format(key, v))

    def print_annotation(self, md):
        annotations = md.get("annotations", None)
        if not annotations:
            return
        for key, value in annotations.items():
            if str(key).startswith("rds-operation"):
                logger.info("{} {}".format(key, value))

    def check_operation_finish(self, name, timeout=600):
        print_logid = False
        while timeout >= 0:
            timeout -= 4
            try:
                res = self.get_operation_by_name(name)
                status = res.get("status")
                md = res.get("metadata")
                if not print_logid:
                    logger.info(
                        "current operation log_id {}".format(
                            status.get('logID')), )
                    print_logid = True
                phase = status.get('phase', "Running")
                if phase == "Running":
                    logger.info("operation {} is running".format(name))
                    self.print_annotation(md)
                elif phase == "Failed":
                    logger.error("operation {} is failed, reason: {}".format(
                        name, status.get('message')))
                    sys.exit(1)
                    return
                elif phase == "Succeeded":
                    logger.info("operation {} is finished".format(name))
                    logger.info("outputParams: {}".format(status.get("outputParams")))
                    self.print_output_params(status.get("outputParams"))
                    self.print_annotation(md)
                    return
                time.sleep(4)
            except Exception as e:
                print(e)
                return


class RDS_API():

    def __init__(self):
        self.url = 'http://mysql.bytecloud:8884/'
        pass

    def __request__(self, action, data):
        url = '{}?Action={}'.format(self.url, action)
        return requests.post(url, json=data).json()

    def create_database(self, ins, db, shard_cnt):
        data = {
            'instance': ins,
            'dbname': db,
            'user': True,
            'priv': 1,
            'is_shard': int(shard_cnt) > 0,
            'shard_cnt': int(shard_cnt)
        }
        return self.__request__('CreateDB', data)

    def create_table(self, **kwargs):
        shard_cnt = int(kwargs.get('shard_cnt'))
        if shard_cnt > 0:
            if not kwargs.get('shard_key') or not kwargs.get(
                    'shard_table') or not kwargs.get('shard_rule'):
                raise Exception(
                    'shard key, table, rule is required when shard_cnt > 0')
        data = {
            'dbname': kwargs.get('dbname'),
            'is_shard': int(shard_cnt) > 0,
            'shard_cnt': int(shard_cnt),
            'sql': kwargs.get('sql'),
            'shard_key': kwargs.get('shard_key'),
            'shard_table': kwargs.get('shard_table'),
            'shard_rule': kwargs.get('shard_rule'),
            'is_ghost': True,
        }
        return self.__request__('UpdateDBDDL', data)


class RDSMGR_API(object):
    def __init__(self):
        self.base_url = "http://rdsmgr.default.svc.storage-cp.org:8089/api/action?Action="
        self.headers = {
            "X-Date": "10.10",
            "Authorization": "aaa",
            "X-Top-User-Id": "0",
            "X-Top-Account-Id": "1",
            "X-Top-Service": "rds_mysql",
            "X-Top-Region": "cn-north-1",
        }

    def call_action(self, action_name, params=None):
        """Call any RDS API action with proper error handling."""
        url = "{}{}".format(self.base_url, action_name)
        params = params or {}
        print(params)

        try:
            response = requests.post(
                url=url,
                headers=self.headers,
                json=params,
                timeout=10
            )
            response.raise_for_status()  # Catch HTTP errors (4xx/5xx)

            # Separate JSON parsing to avoid mixing with Data field checks
            try:
                resp_json = response.json()
            except ValueError:  # JSONDecodeError is subclass of ValueError in Python2
                raise Exception("Invalid JSON response: {}".format(response.text))

            if "Data" in resp_json:
                return resp_json["Data"]
            return resp_json

        except requests.exceptions.RequestException as e:
            raise Exception("Request failed: {}".format(str(e)))
        except Exception as e:
            raise Exception(str(e))


    def parse_param_value(value_str):
        """Return raw string (no JSON parsing for --param values)."""
        return value_str


    def rdsmgr_api_execute(self, type, params):
        try:
            client = RDSMGR_API()
            data = client.call_action(type, params)
        except Exception as e:
            print("Error: {}".format(str(e)))
            exit(1)

        # Handle output fields
        # output_fields = args.output_fields.split(",") if args.output_fields else data.keys()
        for field in data:
            value = data.get(field, "")
            try:
                value_str = json.dumps(value, ensure_ascii=False)
            except:
                value_str = str(value)
            output_param(field, value_str)

        exit(0)