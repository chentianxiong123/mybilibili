import { Modal } from 'antd';
import React from 'react';

export async function alertAsync(content: React.ReactNode | Parameters<typeof Modal.confirm>[0]) {
  if (typeof content === 'object' && content && 'content' in content) {
    return Modal.confirm(content as Parameters<typeof Modal.confirm>[0]);
  }
  return Modal.info({ content });
}

export async function promptAsync(
  options: { title?: string; placeholder?: string },
  defaultValue = '',
) {
  const value = window.prompt(options.title || '', defaultValue);
  return value === null ? defaultValue : value;
}

export async function promptAction(options: {
  title?: string;
  message?: React.ReactNode;
  btnText?: string;
  fn?: () => unknown;
  cancelBtnText?: string;
  cancelBtnPosition?: string;
  secondaryActionText?: string;
  secondaryActionFn?: () => unknown;
}) {
  const result = await new Promise<boolean>(resolve => {
    Modal.confirm({
      title: options.title,
      content: options.message,
      okText: options.btnText || 'OK',
      cancelText: options.cancelBtnText || 'Cancel',
      onOk: async () => {
        if (options.fn) await options.fn();
        resolve(true);
      },
      onCancel: () => resolve(false),
    });
  });
  return result;
}
