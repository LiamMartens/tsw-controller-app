import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
  ReactNode,
} from "react";
import { ErrorBoundary } from "react-error-boundary";
import { useDelayState } from "../hooks";

const ErrorFallback = () => (
  <p className="text-error text-center py-20">An error occured</p>
);
const SuspenseFallback = () => (
  <div className="flex justify-center py-6">
    <span className="loading loading-spinner text-primary"></span>
  </div>
);

export type ModalContentProps<T> = {
  openState: Exclude<T, false>;
  onClose: () => void;
};

export type ModalProps<T> = {
  openState: T | false;
  onClose: () => void;
  Component: (props: ModalContentProps<T>) => ReactNode;
};

export function Modal<T>({ openState, onClose, Component }: ModalProps<T>) {
  const ref = useRef<HTMLDialogElement | null>(null);
  const delayedOpenState = useDelayState(openState);

  const handleRef = (d: HTMLDialogElement | null) => {
    ref.current = d;
  };

  useEffect(() => {
    if (openState) {
      delayedOpenState.setValueInstant(openState);
      ref.current?.showModal();
    } else {
      ref.current?.close();
      delayedOpenState.setValueDelayed(false, 500);
    }
  }, [openState, onClose]);

  return (
    <dialog ref={handleRef} className="modal modal-s">
      <div className="modal-box w-11/12 max-w-5xl">
        <ErrorBoundary fallback={<ErrorFallback />}>
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
    </dialog>
  );
}
