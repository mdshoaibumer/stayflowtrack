"use client";

import { useEffect, useRef, useState } from "react";
import { motion, useMotionValue, animate } from "framer-motion";

interface AnimatedCounterProps {
  value: number;
  prefix?: string;
  suffix?: string;
  duration?: number;
  decimals?: number;
  className?: string;
}

export function AnimatedCounter({
  value,
  prefix = "",
  suffix = "",
  duration = 1,
  decimals = 0,
  className = "",
}: AnimatedCounterProps) {
  const motionValue = useMotionValue(0);
  const [displayValue, setDisplayValue] = useState("0");
  const prevValue = useRef(0);

  useEffect(() => {
    const controls = animate(motionValue, value, {
      duration,
      ease: [0.25, 0.46, 0.45, 0.94],
      onUpdate: (latest) => {
        if (decimals > 0) {
          setDisplayValue(latest.toFixed(decimals));
        } else {
          setDisplayValue(Math.round(latest).toLocaleString("en-IN"));
        }
      },
    });

    prevValue.current = value;
    return () => controls.stop();
  }, [value, duration, decimals, motionValue]);

  return (
    <motion.span
      className={className}
      key={value}
      initial={{ opacity: 0.6, scale: 0.98 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.3 }}
    >
      {prefix}{displayValue}{suffix}
    </motion.span>
  );
}
