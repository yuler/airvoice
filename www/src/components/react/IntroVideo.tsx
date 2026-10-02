import { useRef } from 'react';

interface IntroVideoProps {
  title: string;
  label: string;
  base?: string;
}

function PlayIcon() {
  return (
    <svg className="w-8 h-8 md:w-10 md:h-10 translate-x-[2px]" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M8 5.14v13.72a1 1 0 0 0 1.52.85l10.92-6.86a1 1 0 0 0 0-1.7L9.52 4.29A1 1 0 0 0 8 5.14z" />
    </svg>
  );
}

export default function IntroVideo({ title, label, base = import.meta.env.BASE_URL || '/' }: IntroVideoProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const asset = (name: string) => `${base || '/'}${name}`.replace(/\/+/g, '/');

  const open = () => {
    dialogRef.current?.showModal();
    videoRef.current?.play();
  };

  const close = () => dialogRef.current?.close();

  return (
    <section className="py-16 md:py-20" style={{ background: 'var(--background-primary)' }}>
      <div className="mx-auto max-w-4xl px-4 md:px-6">
        <h2
          className="mb-10 text-center text-2xl font-bold"
          style={{ color: 'var(--primary-text)', letterSpacing: '-0.02em' }}
        >
          {title}
        </h2>
        <button
          type="button"
          onClick={open}
          aria-label={label}
          className="group relative block w-full overflow-hidden rounded-2xl cursor-pointer"
          style={{ border: '1px solid var(--border-default)', background: '#000' }}
        >
          <img
            src={asset('intro-poster.jpg')}
            alt=""
            loading="lazy"
            className="block w-full h-auto aspect-video object-cover transition-transform duration-500 group-hover:scale-[1.02]"
          />
          <span className="absolute inset-0 bg-black/20 transition-colors group-hover:bg-black/30" />
          <span
            className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 flex items-center justify-center w-20 h-20 md:w-28 md:h-28 rounded-full text-white shadow-2xl transition-transform duration-300 group-hover:scale-110 group-active:scale-95"
            style={{ background: '#006efe' }}
          >
            <PlayIcon />
          </span>
        </button>
      </div>

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
    </section>
  );
}
