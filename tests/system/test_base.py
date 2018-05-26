from tcpbeat import BaseTest

import os


class Test(BaseTest):

    def test_base(self):
        """
        Basic test with exiting Tcpbeat normally
        """
        self.render_config_template(
            path=os.path.abspath(self.working_dir) + "/log/*"
        )

        tcpbeat_proc = self.start_beat()
        self.wait_until(lambda: self.log_contains("tcpbeat is running"))
        exit_code = tcpbeat_proc.kill_and_wait()
        assert exit_code == 0
