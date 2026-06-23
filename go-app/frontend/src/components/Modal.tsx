import { Suspense, useEffect, useRef, ReactNode, SyntheticEvent } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { useDelayState } from "../hooks";
import clsx from "clsx";
import { createPortal } from "react-dom";

const ErrorFallback = ({ error }: { error: unknown }) => (
  <div className="alert alert-error">{`An error occured (${error})`}</div>
);
const SuspenseFallback = () => (
  <div className="flex justify-center py-6">
    <span className="loading loading-spinner text-primary"></span>
  </div>
);

export interface ModalCloseReason {}

export type ModalContentProps<T> = {
  openState: Exclude<T, false>;
  onClose: (reason?: ModalCloseReason) => void;
};

export type ModalProps<T> = {
  openState: T | false;
  className?: string;
  onClose: (reason?: ModalCloseReason) => void;
  Component: (props: ModalContentProps<T>) => ReactNode;
};

export function Modal<T>({
  className,
  openState,
  onClose,
  Component,
}: ModalProps<T>) {
  const ref = useRef<HTMLDialogElement | null>(null);
  const delayedOpenState = useDelayState(openState);

  const handleRef = (d: HTMLDialogElement | null) => {
    ref.current = d;
  };

  const handleClose = (e: SyntheticEvent<HTMLDialogElement>) => {
    e.stopPropagation();
    onClose();
  };

  useEffect(() => {
    if (openState) {
      delayedOpenState.setValueInstant(openState);
      ref.current?.showModal();
    } else {
      ref.current?.close();
      delayedOpenState.setValueDelayed(false, 500);
    }
  }, [openState]);

  return createPortal(
    <dialog
      ref={handleRef}
      className={clsx("modal modal-s", className)}
      onClose={handleClose}
    >
      <div className="modal-box w-11/12 max-w-5xl max-h-[calc(90dvh-6rem)]">
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense fallback={<SuspenseFallback />}>
            {delayedOpenState.value !== false && (
              <Component
                openState={delayedOpenState.value as Exclude<T, false>}
                onClose={onClose}
              />
            )}
          </Suspense>
        </ErrorBoundary>
      </div>
    </dialog>,
    document.body,
  );
}
