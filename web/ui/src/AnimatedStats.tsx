import { useEffect, useRef, useState } from 'react';
import './animated-stats.css';

export type StatFormat = {
  prefix?: string;
  suffix?: string;
};

export type StatTuple = readonly [caption: string, value: number, format?: StatFormat];

export type AnimatedStatsProps = {
  items: readonly StatTuple[];
  /** Count-up length in ms. Default 8000. */
  durationMs?: number;
  className?: string;
};

function prefersReducedMotion() {
  return typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

function renderStableNumber(text: string) {
  return [...text].map((char, index) => {
    if (/\d/.test(char)) {
      return (
        <span key={`${index}-${char}`} className="ak-stat-digit">
          {char}
        </span>
      );
    }
    return (
      <span key={`${index}-${char}`} className="ak-stat-char">
        {char}
      </span>
    );
  });
}

type StatCounterProps = {
  value: number;
  active: boolean;
  prefix?: string;
  suffix?: string;
  durationMs: number;
};

function StatCounter({ value, active, prefix = '', suffix = '', durationMs }: StatCounterProps) {
  const [display, setDisplay] = useState(0);
  const finalText = `${prefix}${value.toLocaleString()}${suffix}`;

  useEffect(() => {
    if (!active) {
      setDisplay(0);
      return;
    }

    if (prefersReducedMotion()) {
      setDisplay(value);
      return;
    }

    let frame = 0;
    const start = performance.now();

    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / durationMs);
      // Strong ease-out: decelerates more as it nears the final value
      const eased = 1 - Math.pow(1 - t, 5);
      setDisplay(Math.round(value * eased));
      if (t < 1) frame = requestAnimationFrame(tick);
    };

    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [active, value, durationMs]);

  const text = `${prefix}${display.toLocaleString()}${suffix}`;

  return (
    <div className="ak-stat-value" style={{ minWidth: `${finalText.length}ch` }}>
      {renderStableNumber(text)}
    </div>
  );
}

export function AnimatedStats({ items, durationMs = 8000, className }: AnimatedStatsProps) {
  const rootRef = useRef<HTMLDivElement>(null);
  const [active, setActive] = useState(false);

  useEffect(() => {
    const el = rootRef.current;
    if (!el) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting) {
          setActive(true);
          observer.disconnect();
        }
      },
      { threshold: 0.25 },
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  const rootClassName = ['ak-stats', className].filter(Boolean).join(' ');

  return (
    <div ref={rootRef} className={rootClassName}>
      {items.map(([caption, value, format], index) => (
        <div key={`${caption}-${index}`} className="ak-stats-item">
          <StatCounter
            value={value}
            active={active}
            prefix={format?.prefix}
            suffix={format?.suffix}
            durationMs={durationMs}
          />
          <div className="ak-stat-label">{caption}</div>
        </div>
      ))}
    </div>
  );
}
