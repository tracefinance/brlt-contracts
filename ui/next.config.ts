import type { NextConfig } from "next";

/** @type {import('next').NextConfig} */
const nextConfig: NextConfig = {
  /* config options here */
  distDir: "dist",
  rewrites: async () => {
    return [
      {
        source: '/api/v1/:path*',
        destination: 'http://localhost:8080/api/v1/:path*', // Adjust to your Go API URL
      },
    ];
  },
};

export default nextConfig;
