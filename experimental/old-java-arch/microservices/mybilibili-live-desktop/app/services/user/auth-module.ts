import { IUserAuth } from '.';

export class AuthModule {
  async getAuth(): Promise<IUserAuth | null> {
    return null;
  }
}
