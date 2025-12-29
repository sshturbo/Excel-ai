// MarkdownRenderer - Custom markdown components with syntax highlighting
import { useMemo } from 'react'
import ReactMarkdown, { Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { toast } from 'sonner'
import { cleanTechnicalBlocks } from '@/services/contentCleaner'

/**
 * Custom markdown components with Tailwind styling
 */
export function useMarkdownComponents(): Components {
    return useMemo(() => ({
        // Inline code
        code({ node, className, children, ...props }) {
            const match = /language-(\w+)/.exec(className || '')
            const isInline = !match && !className

            if (isInline) {
                return (
                    <code className="px-1.5 py-0.5 rounded bg-muted text-primary font-mono text-sm" {...props}>
                        {children}
                    </code>
                )
            }

            // Code block with syntax highlighting
            return (
                <div className="relative group my-3">
                    <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                            onClick={() => {
                                navigator.clipboard.writeText(String(children).replace(/\n$/, ''))
                                toast.success('Código copiado!')
                            }}
                            className="px-2 py-1 text-xs bg-muted hover:bg-muted/80 rounded border border-border"
                        >
                            📋 Copiar
                        </button>
                    </div>
                    {match && (
                        <div className="text-xs text-muted-foreground px-3 py-1 bg-muted/50 border-b border-border rounded-t">
                            {match[1]}
                        </div>
                    )}
                    <SyntaxHighlighter
                        style={oneDark}
                        language={match?.[1] || 'text'}
                        PreTag="div"
                        customStyle={{
                            margin: 0,
                            borderRadius: match ? '0 0 0.5rem 0.5rem' : '0.5rem',
                            fontSize: '0.85rem',
                        }}
                    >
                        {String(children).replace(/\n$/, '')}
                    </SyntaxHighlighter>
                </div>
            )
        },
        // Tables
        table({ children }) {
            return (
                <div className="overflow-x-auto my-3">
                    <table className="min-w-full border border-border rounded-lg overflow-hidden">
                        {children}
                    </table>
                </div>
            )
        },
        thead({ children }) {
            return <thead className="bg-muted/50">{children}</thead>
        },
        th({ children }) {
            return <th className="px-3 py-2 text-left text-sm font-semibold border-b border-border">{children}</th>
        },
        td({ children }) {
            return <td className="px-3 py-2 text-sm border-b border-border/50">{children}</td>
        },
        // Lists
        ul({ children }) {
            return <ul className="list-disc list-inside space-y-1 my-2 ml-2">{children}</ul>
        },
        ol({ children }) {
            return <ol className="list-decimal list-inside space-y-1 my-2 ml-2">{children}</ol>
        },
        li({ children }) {
            return <li className="text-sm">{children}</li>
        },
        // Headings
        h1({ children }) {
            return <h1 className="text-xl font-bold mt-4 mb-2 text-primary">{children}</h1>
        },
        h2({ children }) {
            return <h2 className="text-lg font-bold mt-3 mb-2 text-primary">{children}</h2>
        },
        h3({ children }) {
            return <h3 className="text-base font-semibold mt-2 mb-1">{children}</h3>
        },
        // Paragraphs
        p({ children }) {
            return <p className="my-2 leading-relaxed">{children}</p>
        },
        // Links
        a({ href, children }) {
            return (
                <a
                    href={href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline"
                >
                    {children}
                </a>
            )
        },
        // Blockquote
        blockquote({ children }) {
            return (
                <blockquote className="border-l-4 border-primary/50 pl-4 my-3 italic text-muted-foreground bg-muted/20 py-2 rounded-r">
                    {children}
                </blockquote>
            )
        },
        // Horizontal rule
        hr() {
            return <hr className="my-4 border-border" />
        },
        // Strong/Bold
        strong({ children }) {
            return <strong className="font-semibold text-primary">{children}</strong>
        },
        // Emphasis/Italic
        em({ children }) {
            return <em className="italic">{children}</em>
        },
    }), [])
}

interface MarkdownRendererProps {
    content: string
}

/**
 * Renders markdown content with custom styling
 */
export function MarkdownRenderer({ content }: MarkdownRendererProps) {
    const markdownComponents = useMarkdownComponents()

    // First clean technical blocks (including :::thinking)
    const cleanedContent = cleanTechnicalBlocks(content)

    if (!cleanedContent) {
        return null
    }

    return (
        <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
            {cleanedContent}
        </ReactMarkdown>
    )
}
