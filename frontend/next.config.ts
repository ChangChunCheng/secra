import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  output: 'export', // Enable static export
  images: {
    unoptimized: true, // Static export doesn't support Next.js default Image Optimization
  },
};

export default nextConfig;
