export interface ApiErr {
  msg: string;
  code: string;
  data?: object;
  request_id: string;
}

export const isApiErr = (err: unknown): err is ApiErr =>
  err && typeof err === 'object' ? 'msg' in err && 'code' in err && 'request_id' in err : false;

interface ApiVersion {
  env: string;
  last_commit: string;
  build_time: string;
  time: string;
}

export const toVersion = ({ time, build_time, last_commit, env }: ApiVersion) => {
  return {
    env,
    lastCommit: last_commit,
    buildTime: new Date(build_time),
    time: new Date(time),
  };
};

export type Version = ReturnType<typeof toVersion>;
