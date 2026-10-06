import { useEffect, useRef, useState } from "react";

function prefersReducedMotion(): boolean {
  return typeof window.matchMedia === "function" && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

/**
 * Smoothly animates a number towards `target` using requestAnimationFrame.
 * The frame is always cancelled on unmount / target change (no leaks) and the
 * value snaps instantly for users who prefer reduced motion.
 */
export function useAnimatedNumber(target: number, durationMs = 550): number {
  const [displayed, setDisplayed] = useState(target);
  const displayedRef = useRef(target);
  const frameRef = useRef<number | null>(null);

  useEffect(() => {
    const from = displayedRef.current;
    if (from === target) {
      return;
    }
    if (prefersReducedMotion()) {
      displayedRef.current = target;
      setDisplayed(target);
      return;
    }

    const start = performance.now();
    const step = (now: number) => {
      const progress = Math.min(1, (now - start) / durationMs);
      const eased = 1 - Math.pow(1 - progress, 3);
      const value = Math.round(from + (target - from) * eased);
      displayedRef.current = value;
      setDisplayed(value);
      if (progress < 1) {
        frameRef.current = requestAnimationFrame(step);
      }
    };
    frameRef.current = requestAnimationFrame(step);

    return () => {
      if (frameRef.current !== null) {
        cancelAnimationFrame(frameRef.current);
      }
    };
  }, [target, durationMs]);

  return displayed;
}
