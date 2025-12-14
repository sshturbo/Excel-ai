// ChartViewer component - Display charts from spreadsheet data
import { Bar, Line, Pie } from 'react-chartjs-2'
import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    BarElement,
    LineElement,
    PointElement,
    Title,
    Tooltip,
    Legend,
    ArcElement
} from 'chart.js'

// Register Chart.js components
ChartJS.register(
    CategoryScale,
    LinearScale,
    BarElement,
    LineElement,
    PointElement,
    Title,
    Tooltip,
    Legend,
    ArcElement
)

interface ChartViewerProps {
    chartType: 'bar' | 'line' | 'pie'
    chartData: any
}

export function ChartViewer({ chartType, chartData }: ChartViewerProps) {
    const chartOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: { labels: { color: '#e6edf3' } },
            title: { display: true, text: 'Visualização', color: '#e6edf3' }
        },
        scales: chartType !== 'pie' ? {
            x: { ticks: { color: '#8b949e' }, grid: { color: '#21262d' } },
            y: { ticks: { color: '#8b949e' }, grid: { color: '#21262d' } }
        } : undefined
    }

    return (
        <div className="flex-1 flex items-center justify-center p-6">
            <div className="w-full max-h-96">
                {chartType === 'bar' && <Bar data={chartData} options={chartOptions} />}
                {chartType === 'line' && <Line data={chartData} options={chartOptions} />}
                {chartType === 'pie' && <Pie data={chartData} options={chartOptions} />}
            </div>
        </div>
    )
}
