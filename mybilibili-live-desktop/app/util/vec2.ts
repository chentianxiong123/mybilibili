type TVecConstructorArgs = [number, number] | [IVec2] | [];

export class Vec2 {
  x: number;
  y: number;

  constructor(...args: TVecConstructorArgs) {
    if (args.length === 0) {
      this.x = 0;
      this.y = 0;
    } else if (typeof args[0] === 'number') {
      this.x = args[0];
      this.y = args[1];
    } else {
      this.x = args[0].x;
      this.y = args[0].y;
    }
  }

  clone() {
    return new Vec2(this.x, this.y);
  }

  add(v: IVec2) {
    return new Vec2(this.x + v.x, this.y + v.y);
  }

  sub(v: IVec2) {
    return new Vec2(this.x - v.x, this.y - v.y);
  }

  multiplyScalar(value: number) {
    return new Vec2(this.x * value, this.y * value);
  }

  divideScalar(value: number) {
    return new Vec2(this.x / value, this.y / value);
  }

  length() {
    return Math.sqrt(this.x * this.x + this.y * this.y);
  }

  normalize() {
    const length = this.length();
    return length ? this.divideScalar(length) : new Vec2();
  }
}

export function v2(): Vec2;
export function v2(x: number, y: number): Vec2;
export function v2(model: IVec2): Vec2;
export function v2(...args: TVecConstructorArgs) {
  return new Vec2(...args);
}
