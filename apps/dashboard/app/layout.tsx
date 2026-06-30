import type { Metadata } from 'next'
import { Inter, JetBrains_Mono } from 'next/font/google'

import { Toaster } from '@/components/ui/toaster'
import '@/app/globals.css'

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
})

const jetbrainsMono = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-jetbrains-mono',
})

export const metadata: Metadata = {
  title: 'Bolt Monitor Dashboard',
  description: 'Operator dashboard for monitor management and runtime status inspection.',
}

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html className="dark" lang="en">
      <body className={`${inter.variable} ${jetbrainsMono.variable} font-sans`}>
        <a
          className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-md focus:border focus:border-primary/40 focus:bg-surface focus:px-3 focus:py-2 focus:text-sm focus:font-semibold focus:text-primary"
          href="#main-content"
        >
          Skip to main content
        </a>
        {children}
        <Toaster />
      </body>
    </html>
  )
}
