interface ApiUser {
  _id: string;
  name: string;
  email: string;
  type: string;
  status: string;
}

export const toUser = ({ _id, name, email, type, status }: ApiUser) => ({
  id: _id,
  name,
  email,
  type,
  status,
});

export type User = ReturnType<typeof toUser>;
