import React, {useEffect, useState} from 'react';
import {Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis} from "recharts";
import axios from "axios";
import {Modal, ModalBody, ModalContent, ModalHeader} from "@heroui/react";



export const FansChart = function FansChart(props) {
    var color = 'success'
    const [data, setData] = useState([]);
    const [yDomain, setYDomain] = useState(['auto', 'auto']);

    useEffect(() => {
        if (props.data.length > 0) {
            var cpy = []
            props.data.forEach((d,i) => {
                cpy.push(d)
                cpy[i].Time = new Date(d.CreatedAt).toLocaleDateString();
            })
            setData(cpy);
            const fansValues = cpy.map(d => d.Fans);
            const minFans = Math.min(...fansValues);
            const maxFans = Math.max(...fansValues);
            const range = maxFans - minFans;


            let padding;
            if (range === 0) {
                padding = minFans * 0.01;
            } else if (range / minFans < 0.01) {
                padding = range * 2;
            } else {
                padding = range * 0.1;
            }

            setYDomain([minFans - padding, maxFans + padding]);
        }
    }, [props.data]);
    const CustomTooltip = ({ active, payload, label }) => {
        if (active && payload && payload.length) {
            return (
                <div className="bg-white p-2 border border-gray-200 rounded shadow">
                    <p className="text-sm">{`Date: ${label}`}</p>
                    <p className="text-sm font-semibold">{`Value: ${payload[0].value.toLocaleString()}`}</p>
                </div>
            );
        }
        return null;
    };
    const formatYAxis = (value) => {
        if (value >= 1000000) {
            return `${(value / 1000000).toFixed(1)}M`;
        } else if (value >= 1000) {
            return `${(value / 1000).toFixed(1)}K`;
        }
        return value.toLocaleString();
    };
    return (
        <ResponsiveContainer className={'h-full'}>
            <AreaChart
                accessibilityLayer
                data={data}
                height={300}
                margin={{
                    left: 10,
                    right: 10,
                    top: 10,
                    bottom: 10
                }}
            >
                <defs>
                    <linearGradient id="colorGradient" x1="0" x2="0" y1="0" y2="1">
                        <stop
                            offset="10%"
                            stopColor={`hsl(var(--heroui-${color}-500))`}
                            stopOpacity={0.3}
                        />
                        <stop
                            offset="100%"
                            stopColor={`hsl(var(--heroui-${color}-100))`}
                            stopOpacity={0.1}
                        />
                    </linearGradient>
                </defs>
                <CartesianGrid
                    stroke="hsl(var(--heroui-default-200))"
                    strokeDasharray="3 3"
                    vertical={false}
                />
                <XAxis
                    axisLine={false}
                    dataKey="Time"
                    style={{fontSize: "var(--heroui-font-size-tiny)"}}
                    tickLine={false}
                    tickMargin={5}
                />
                <YAxis
                    domain={yDomain}
                    axisLine={false}
                    tickLine={false}
                    tickMargin={8}
                    tickFormatter={formatYAxis}
                    tickCount={6}
                    style={{ fontSize: "var(--heroui-font-size-tiny)" }}
                />
                <Tooltip content={<CustomTooltip />} />
                <Area
                    type="monotone"
                    dataKey="Fans"
                    stroke={`hsl(var(--heroui-${color}))`}
                    strokeWidth={2}
                    fill="url(#colorGradient)"
                    animationDuration={1000}
                    animationEasing="ease"
                    activeDot={{
                        stroke: `hsl(var(--heroui-${color}))`,
                        strokeWidth: 2,
                        fill: "hsl(var(--heroui-background))",
                        r: 5,
                    }}
                />
            </AreaChart>
        </ResponsiveContainer >
    );
}

export const GuardChart = function GuardChart(props) {
    var color = 'success'
    const [data, setData] = useState([]);
    const [yDomain, setYDomain] = useState(['auto', 'auto']);

    useEffect(() => {
        if (props.data.length > 0) {
            var cpy = []
            props.data.forEach((d,i) => {
                cpy.push(d)
                cpy[i].Time = new Date(d.UpdatedAt).toLocaleDateString();
            })
            setData(cpy);
        }
    }, [props.data]);
    const CustomTooltip = ({ active, payload, label }) => {
        if (active && payload && payload.length) {
            return (
                <div className="bg-white p-2 border border-gray-200 rounded shadow">
                    <p className="text-sm">{`Date: ${label}`}</p>
                    <p className="text-sm font-semibold">{`Value: ${payload[0].value.toLocaleString()}`}</p>
                </div>
            );
        }
        return null;
    };
    return (
        <ResponsiveContainer className={'h-full'}>
            <AreaChart
                accessibilityLayer
                data={data}
                margin={{
                    left: 10,
                    right: 10,
                    top: 10,
                    bottom: 10
                }}
                onClick={(e, payload) => {
                    console.log(e)
                    console.log(payload);
                    props.onClick(parseInt(e.activeIndex))
                }}
            >
                <defs>
                    <linearGradient id="colorGradient" x1="0" x2="0" y1="0" y2="1">
                        <stop
                            offset="10%"
                            stopColor={`hsl(var(--heroui-${color}-500))`}
                            stopOpacity={0.3}
                        />
                        <stop
                            offset="100%"
                            stopColor={`hsl(var(--heroui-${color}-100))`}
                            stopOpacity={0.1}
                        />
                    </linearGradient>
                </defs>
                <CartesianGrid
                    stroke="hsl(var(--heroui-default-200))"
                    strokeDasharray="3 3"
                    vertical={false}
                />
                <XAxis
                    axisLine={false}
                    dataKey="Time"
                    style={{fontSize: "var(--heroui-font-size-tiny)"}}
                    tickLine={false}
                    tickMargin={5}
                />
                <YAxis
                    domain={yDomain}
                    axisLine={false}
                    tickLine={false}
                    tickMargin={8}
                    tickFormatter={(value) => Math.round(value).toString()}
                    tickCount={6}
                    style={{ fontSize: "var(--heroui-font-size-tiny)" }}
                />
                <Tooltip content={<CustomTooltip />} />
                <Area
                    type="monotone"
                    dataKey="Guard"
                    stroke={`hsl(var(--heroui-${color}))`}
                    strokeWidth={2}
                    fill="url(#colorGradient)"
                    animationDuration={1000}
                    animationEasing="ease"
                    activeDot={{
                        stroke: `hsl(var(--heroui-${color}))`,
                        strokeWidth: 2,
                        fill: "hsl(var(--heroui-background))",
                        r: 5,
                    }}
                />
            </AreaChart>
        </ResponsiveContainer>
    );
}

export const LiveMessageChart = function LiveMessageChart(props) {
    var color = 'success'
    const [data, setData] = useState([]);
    const [yDomain, setYDomain] = useState(['auto', 'auto']);

    useEffect(() => {
        props.id && axios.get("/api/minute?id=" + props.id).then(data => {
            if (!data.data.data) {
                props.onClose()
                return
            }
            setData(data.data.data);
            const fansValues = data.data.data.map(d => d.Count);
            const minFans = Math.min(...fansValues);
            const maxFans = Math.max(...fansValues);
            const range = maxFans - minFans;


            let padding;
            if (range === 0) {
                padding = minFans * 0.01;
            } else if (range / minFans < 0.01) {
                padding = range * 2;
            } else {
                padding = range * 0.1;
            }

            setYDomain([minFans - padding, maxFans + padding]);
        })

    }, []);
    const CustomTooltip = ({ active, payload, label }) => {
        if (active && payload && payload.length) {
            return (
                <div className="bg-white p-2 border border-gray-200 rounded shadow">
                    <p className="text-sm">{`Date: ${label}`}</p>
                    <p className="text-sm font-semibold">{`Value: ${payload[0].value.toLocaleString()}`}</p>
                </div>
            );
        }
        return null;
    };
    return (
        <Modal isOpen={true} size={'3xl'} onClose={props.onClose}>
            <ModalContent  className={'w-[1200px] h-[600px]'}>
                <ModalBody >
                    <ModalHeader>
                        <h2>Messages  Trending</h2>
                    </ModalHeader>
                    <ResponsiveContainer  >
                        <AreaChart
                            accessibilityLayer
                            data={data}
                            margin={{
                                left: 10,
                                right: 10,
                                top: 10,
                                bottom: 10
                            }}
                        >
                            <defs>
                                <linearGradient id="colorGradient" x1="0" x2="0" y1="0" y2="1">
                                    <stop
                                        offset="10%"
                                        stopColor={`hsl(var(--heroui-${color}-500))`}
                                        stopOpacity={0.3}
                                    />
                                    <stop
                                        offset="100%"
                                        stopColor={`hsl(var(--heroui-${color}-100))`}
                                        stopOpacity={0.1}
                                    />
                                </linearGradient>
                            </defs>
                            <CartesianGrid
                                stroke="hsl(var(--heroui-default-200))"
                                strokeDasharray="3 3"
                                vertical={false}
                            />
                            <XAxis
                                axisLine={false}
                                dataKey="MinuteTime"
                                style={{fontSize: "var(--heroui-font-size-tiny)"}}
                                tickLine={false}
                                tickMargin={5}
                                tickFormatter={(value) => new Date(value).toLocaleTimeString()}
                            />
                            <YAxis
                                domain={yDomain}
                                axisLine={false}
                                tickLine={false}
                                tickMargin={8}
                                tickFormatter={(value) => Math.round(value).toString()}
                                tickCount={6}
                                style={{ fontSize: "var(--heroui-font-size-tiny)" }}
                            />
                            <Tooltip content={<CustomTooltip />} />
                            <Area
                                type="monotone"
                                dataKey="RecordCount"
                                stroke={`hsl(var(--heroui-${color}))`}
                                strokeWidth={2}
                                fill="url(#colorGradient)"
                                animationDuration={1000}
                                animationEasing="ease"
                                activeDot={{
                                    stroke: `hsl(var(--heroui-${color}))`,
                                    strokeWidth: 2,
                                    fill: "hsl(var(--heroui-background))",
                                    r: 5,
                                }}
                            />
                        </AreaChart>
                    </ResponsiveContainer >
                </ModalBody>
            </ModalContent>

        </Modal>

    );
}
