import { useRef } from 'react';

interface IntroVideoProps {
  label: string;
  base: string;
}

function PlayIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M8 5.14v13.72a1 1 0 0 0 1.52.85l10.92-6.86a1 1 0 0 0 0-1.7L9.52 4.29A1 1 0 0 0 8 5.14z" />
    </svg>
  );
}

export default function IntroVideo({ label, base }: IntroVideoProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const asset = (name: string) => `${base || '/'}${name}`.replace(/\/+/g, '/');

  const open = () => {
    dialogRef.current?.showModal();
    videoRef.current?.play();
  };

  const close = () => dialogRef.current?.close();

  return (
    <>
      <button
        type="button"
        onClick={open}
        className="inline-flex items-center justify-center gap-1.5 rounded-full border px-4 text-xs font-bold transition-colors hover:bg-black/5 dark:hover:bg-white/10 cursor-pointer"
        style={{ borderColor: 'var(--border-default)', color: 'var(--primary-text)', background: 'transparent', height: '40px', minWidth: '136px' }}
      >
        <PlayIcon />
        {label}
      </button>

      <dialog
        ref={dialogRef}
        className="lightbox-dialog p-0 border-none rounded-2xl bg-transparent outline-none max-w-[90vw] max-h-[90vh] shadow-2xl overflow-visible"
        onClick={(e) => e.target === dialogRef.current && close()}
        onClose={() => videoRef.current?.pause()}
      >
        <div className="relative">
          <video
            ref={videoRef}
            src={asset('intro.mp4')}
            poster={asset('intro-poster.jpg')}
            className="block w-[90vw] max-w-[1280px] max-h-[85vh] rounded-2xl bg-black"
            controls
            playsInline
            preload="none"
          />
          <button
            className="absolute -right-8 -top-8 flex items-center justify-center w-10 h-10 text-white/80 hover:text-white hover:scale-110 active:scale-95 transition-all text-2xl font-normal cursor-pointer border-none outline-none bg-transparent z-10"
            onClick={close}
            aria-label="Close"
          >
            ✕
          </button>
        </div>
      </dialog>
    </>
  );
}
