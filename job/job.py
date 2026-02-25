#!/usr/bin/env python3
# encoding=utf-8
from python_lib.rds_operation import RdsOperation
import sys
if __name__ == "__main__":
    rds = RdsOperation()
    try:
        source_clusterId = sys.argv[1] if len(sys.argv) >= 2 else ""
        new_password = sys.argv[2] if len(sys.argv) >= 3 else ""
        name = rds.create_operation_and_print_result(
            'BDGateGetOrModifyEncryptedPassword', {
                'sourceClusterId': source_clusterId,
                'newPassword': new_password
            })
        rds.check_operation_finish(name, timeout=900)
    except Exception as e:
        print(e)
        sys.exit(1)