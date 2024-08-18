from echobox.tool import template
from echobox.tool import functocli
from echobox.tool import dockerutil
from echobox.app.devops import DevOpsApp

APP_NAME = 'ikuai-exporter'
image = 'kiuber/ikuai-exporter'


class App(DevOpsApp):

    def __init__(self):
        DevOpsApp.__init__(self, APP_NAME)

    def build_image(self, buildx=True, push=False):
        build_params = f'-t {image} {self.root_dir}'
        if buildx:
            self.shell_run('docker buildx create --use --name multi-arch-builder', exit_on_error=False)
            cmd = f'docker buildx build --platform linux/amd64,linux/arm64 {build_params}'
            if push:
                cmd += f' --push'
        else:
            cmd = f'docker build {build_params}'
        self.shell_run(cmd)

    def restart(self, ikuai_ip, username, password, metrics_port=12695, pushgateway_url='', pushgateway_crontab='*/15 * * * * *', pushgateway_job='ikuai', debug=False):
        container = f'{self.app_name}-{ikuai_ip}'
        self.stop_container(container, timeout=1)
        self.remove_container(container, force=True)

        ports = [f'{metrics_port}:9090'] if metrics_port else []

        envs = [
            f'IK_URL=http://{ikuai_ip}',
            f'IK_USER={username}',
            f'IK_PWD={password}',
            f'DEBUG={debug}',
        ]
        if pushgateway_url:
            envs.append(f'PG_URL={pushgateway_url}')
            envs.append(f'PG_JOB={pushgateway_job}')
            if pushgateway_crontab:
                envs.append(f'PG_CRONTAB="{pushgateway_crontab}"')

        args = dockerutil.base_docker_args(container_name=container, ports=ports, envs=envs)

        cmd_data = {'image': image, 'args': args}
        cmd = template.render_str('docker run -d --restart always {{ args }} {{ image }}', cmd_data)
        self.shell_run(cmd)


if __name__ == '__main__':
    functocli.run_app(App)
