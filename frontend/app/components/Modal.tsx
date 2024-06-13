import { MouseEventHandler, ReactNode, useRef } from "react";

export default function Modal({
  isOpen,
  onClose = () => {},
  children,
}: {
  isOpen: boolean;
  onClose: () => void;
  children?: ReactNode;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  if (isOpen) {
    dialog.current?.showModal();
  } else {
    dialog.current?.close();
  }
  const onClick: MouseEventHandler<HTMLDialogElement> = (ev) => {
    const rect = ev.currentTarget.getBoundingClientRect();
    if (!rect) return;
    if (
      ev.clientX < rect.left ||
      ev.clientX > rect.right ||
      ev.clientY < rect.top ||
      ev.clientY > rect.bottom
    ) {
      dialog.current?.close();
    }
  };

  return (
    <dialog
      className="w-fit rounded-lg p-6 surface border backdrop:bg-transparent"
      ref={dialog}
      onMouseDown={onClick}
      onClose={onClose}
    >
      {children}
    </dialog>
  );
}
