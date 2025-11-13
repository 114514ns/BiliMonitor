import React, { useState } from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend, Sector } from 'recharts';

const HEROUI_COLORS = [
    'hsl(var(--heroui-primary))',
    'hsl(var(--heroui-success))',
    'hsl(var(--heroui-warning))',
    'hsl(var(--heroui-danger))',
    'hsl(var(--heroui-secondary))',
    'hsl(var(--heroui-primary-300))',
    'hsl(var(--heroui-success-300))',
    'hsl(var(--heroui-warning-300))',
];

const renderActiveShape = (props) => {
    const RADIAN = Math.PI / 180;
    const {
        cx, cy, midAngle, innerRadius, outerRadius, startAngle, endAngle,
        fill, payload, percent, value
    } = props;
    const sin = Math.sin(-RADIAN * midAngle);
    const cos = Math.cos(-RADIAN * midAngle);
    const sx = cx + (outerRadius + 10) * cos;
    const sy = cy + (outerRadius + 10) * sin;
    const mx = cx + (outerRadius + 30) * cos;
    const my = cy + (outerRadius + 30) * sin;
    const ex = mx + (cos >= 0 ? 1 : -1) * 22;
    const ey = my;
    const textAnchor = cos >= 0 ? 'start' : 'end';

    return (
        <g>
            <text
                x={cx}
                y={cy}
                dy={8}
                textAnchor="middle"
                fill={fill}
                className="text-2xl font-semibold"
            >
                {payload.name}
            </text>
            <Sector
                cx={cx}
                cy={cy}
                innerRadius={innerRadius}
                outerRadius={outerRadius}
                startAngle={startAngle}
                endAngle={endAngle}
                fill={fill}
            />
            <Sector
                cx={cx}
                cy={cy}
                startAngle={startAngle}
                endAngle={endAngle}
                innerRadius={outerRadius + 6}
                outerRadius={outerRadius + 10}
                fill={fill}
            />
            <path d={`M${sx},${sy}L${mx},${my}L${ex},${ey}`} stroke={fill} fill="none" />
            <circle cx={ex} cy={ey} r={2} fill={fill} stroke="none" />
            <text
                x={ex + (cos >= 0 ? 1 : -1) * 12}
                y={ey}
                textAnchor={textAnchor}
                fill="hsl(var(--heroui-default-700))"
                className="text-sm"
            >
                {`${value.toLocaleString()}`}
            </text>
            <text
                x={ex + (cos >= 0 ? 1 : -1) * 12}
                y={ey}
                dy={18}
                textAnchor={textAnchor}
                fill="hsl(var(--heroui-default-500))"
                className="text-xs"
            >
                {`(${(percent * 100).toFixed(2)}%)`}
            </text>
        </g>
    );
};

const CustomTooltip = ({ active, payload }) => {
    if (active && payload && payload.length) {
        const data = payload[0];
        return (
            <div className="bg-white/95 backdrop-blur-md p-3 rounded-lg shadow-lg border border-default-200">
                <p className="text-sm font-medium text-default-700">{data.name}</p>
                <p className="text-lg font-semibold" style={{ color: data.payload.fill }}>
                    {data.value.toLocaleString()}
                </p>
                <p className="text-xs text-default-500">
                    {((data.value / data.payload.total) * 100).toFixed(2)}%
                </p>
            </div>
        );
    }
    return null;
};

const CustomLegend = ({ payload }) => {
    return (
        <ul className="flex flex-wrap justify-center gap-3 mt-4">
            {payload.map((entry, index) => (
                <li key={`item-${index}`} className="flex items-center gap-2">
          <span
              className="w-3 h-3 rounded-full"
              style={{ backgroundColor: entry.color }}
          />
                    <span className="text-sm text-default-700">{entry.value}</span>
                </li>
            ))}
        </ul>
    );
};

export const HeroUIPieChart = ({
                                   data,
                                   dataKey = 'value',
                                   nameKey = 'name',
                                   height =isMobile()?400:400*(innerWidth/1366),
                                   showLegend = true,
                                   showTooltip = true,
                                   innerRadius = 0,
                                   outerRadius = 120,
                                   startAngle = 90,
                                   endAngle = -270,
                                   colors = HEROUI_COLORS,
                                   className = '',
                                   onSegmentClick,
                                    width = ''
                               }) => {
    const [activeIndex, setActiveIndex] = useState(null);
    const total = data.reduce((sum, item) => sum + item[dataKey], 0);
    const processedData = data.map(item => ({
        ...item,
        total
    }));

    const onPieEnter = (_, index) => {
        setActiveIndex(index);
    };

    const onPieLeave = () => {
        setActiveIndex(null);
    };

    const handleClick = (data, index) => {
        if (onSegmentClick) {
            onSegmentClick(data, index);
        }
    };

    return (
        <div className={` ${className}`}>
            <ResponsiveContainer width={isMobile()?vwToPx(90):vwToPx(35)} height={height}>
                <PieChart>
                    <Pie
                        activeIndex={activeIndex}
                        activeShape={renderActiveShape}
                        data={processedData}
                        cx="50%"
                        cy="50%"
                        innerRadius={innerRadius}
                        outerRadius={outerRadius}
                        startAngle={startAngle}
                        endAngle={endAngle}
                        dataKey={dataKey}
                        nameKey={nameKey}
                        onMouseEnter={onPieEnter}
                        onMouseLeave={onPieLeave}
                        onClick={handleClick}
                    >
                        {data.map((entry, index) => (
                            <Cell
                                key={`cell-${index}`}
                                fill={colors[index % colors.length]}
                                className="cursor-pointer hover:opacity-80 transition-opacity"
                            />
                        ))}
                    </Pie>
                    {showTooltip && <Tooltip content={<CustomTooltip />} />}
                    {showLegend && <Legend content={<CustomLegend />} />}
                </PieChart>
            </ResponsiveContainer>
        </div>
    );
};