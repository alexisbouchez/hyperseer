import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  output: 'standalone',
  async rewrites() {
    return [
      { source: '/api/:path*', destination: `${process.env.API_URL ?? 'http://hyperseer:7777'}/:path*` },
    ]
  },
}

export default nextConfig
