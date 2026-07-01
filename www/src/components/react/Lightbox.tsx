import React, { useRef } from 'react';

interface LightboxProps {
  src: string;
  alt?: string;
  className?: string;
  style?: React.CSSProperties;
}

export default function Lightbox({ src, alt, className, style }: LightboxProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);

  const openDialog = () => {
    if (dialogRef.current) {
      dialogRef.current.showModal();
    }
  };

  const closeDialog = () => {
    if (dialogRef.current) {
      dialogRef.current.close();
    }
  };

  // Close when clicking on the backdrop area
  const handleDialogClick = (e: React.MouseEvent<HTMLDialogElement>) => {
    if (e.target === dialogRef.current) {
      closeDialog();
    }
  };

  return (
    <>
      <style>{`
        .lightbox-dialog {
          opacity: 0;
          transition: all 0.25s ease;
        }
        .lightbox-dialog[open] {
          opacity: 1;
          position: fixed;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          margin: 0;
          animation: dialog-zoom-in 0.25s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
        }
        .lightbox-dialog[open]::backdrop {
          animation: dialog-fade-in 0.25s ease-out forwards;
        }
        @keyframes dialog-zoom-in {
          from {
            opacity: 0;
            transform: translate(-50%, -50%) scale(0.92);
          }
          to {
            opacity: 1;
            transform: translate(-50%, -50%) scale(1);
          }
        }
        @keyframes dialog-fade-in {
          from {
            background-color: rgba(0, 0, 0, 0);
            backdrop-filter: blur(0px);
            -webkit-backdrop-filter: blur(0px);
          }
          to {
            background-color: rgba(0, 0, 0, 0.8);
            backdrop-filter: blur(6px);
            -webkit-backdrop-filter: blur(6px);
          }
        }
      `}</style>

      <img
        src={src}
        alt={alt}
        className={`${className || ''} cursor-zoom-in transition-transform duration-300 hover:scale-[1.02]`}
        style={style}
        onClick={openDialog}
      />

      <dialog
        ref={dialogRef}
        className="lightbox-dialog p-0 border-none rounded-2xl bg-transparent outline-none max-w-[90vw] max-h-[90vh] shadow-2xl overflow-visible"
        onClick={handleDialogClick}
      >
        <div className="relative inline-block p-0 m-0 overflow-visible">
          <img
            src={src}
            alt={alt || 'Zoomed'}
            className="max-w-[90vw] max-h-[85vh] object-contain rounded-2xl select-none cursor-zoom-out block border-none outline-none"
            onClick={closeDialog}
          />
          {/* Close button outside the top-right corner of the image, transparent background */}
          <button
            className="absolute -right-8 -top-8 flex items-center justify-center w-10 h-10 text-white/80 hover:text-white hover:scale-110 active:scale-95 transition-all text-2xl font-normal cursor-pointer border-none outline-none bg-transparent z-10"
            onClick={closeDialog}
            aria-label="Close"
          >
            ✕
          </button>
        </div>
      </dialog>
    </>
  );
}
