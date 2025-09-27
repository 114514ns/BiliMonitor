import React, {useEffect} from 'react';
import axios from "axios";
import {Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis} from "recharts";

function StatusPage(props) {

    function formatNumber(num, {
        decimals = 1,
        suffix = '',
    } = {}) {
        if (num === null || num === undefined || isNaN(num)) return '0';

        const units = ['', 'K', 'M', 'G', 'T', 'P'];
        const tier = Math.floor(Math.log10(Math.abs(num)) / 3);

        if (tier === 0) return num + suffix;

        const unit = units[tier];
        const scaled = num / Math.pow(10, tier * 3);
        return scaled.toFixed(decimals) + unit + suffix;
    }


    const [overview, setOverview] = React.useState({});

    const color = 'success'
    const [msg, setMsg] = React.useState([]);
    useEffect(() => {
        var prefix = 'api'
        if (import.meta.env.PROD) {
            prefix = ""
        }
        let ws = new WebSocket(`${prefix}/status`);
        ws.onclose = () => {
            console.log('WebSocket disconnected');
        }
        ws.onmessage = e => {
            setOverview(JSON.parse(e.data))
        }
        var intervalId = 0
        ws.onopen = () => {
            intervalId = setInterval(() => {
                ws.send('鸡你太美')
            }, 2000)
        }

        return () => clearInterval(intervalId)
    }, [])
    useEffect(() => {
        axios.get('/api/chart/msg').then(res => {
            setMsg(res.data.data);
        })
    }, [])

    const contentMap = ['24小时弹幕', '月弹幕', '6月弹幕']
    return (

        <div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm">
                <div
                    className="rounded-xl bg-pink-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">每分钟弹幕<br/><span
                    className="font-semibold">{overview.Message1}</span>
                </div>
                <div
                    className="rounded-xl bg-gray-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">每5分钟弹幕<br/>
                    <span className="font-semibold">{overview.Message5}</span>
                </div>
                <div
                    className="rounded-xl bg-green-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">每小时弹幕<br/>
                    <span
                        className="font-semibold">{overview.MessageHour}</span>
                </div>

            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm mt-4">
                <div
                    className="rounded-xl bg-orange-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">24小时弹幕<br/>
                    <span
                        className="font-semibold">{overview.MessageDaily}</span>
                </div>
                <div
                    className="rounded-xl bg-blue-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">总弹幕<br/>
                    <span
                        className="font-semibold">{overview.TotalMessages}</span>
                </div>
                <div
                    className="rounded-xl bg-blue-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">启动时间<br/>
                    <span
                        className="font-semibold">{overview.LaunchedAt}</span>
                </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm mt-4">
                <div
                    className="rounded-xl bg-orange-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">HTTP请求书<br/>
                    <span
                        className="font-semibold">{formatNumber(overview.Requests)}</span>
                </div>
                <div
                    className="rounded-xl bg-yellow-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">HTTP字节数<br/>
                    <span
                        className="font-semibold">{formatNumber(overview.HTTPBytes)}</span>
                </div>
                <div
                    className="rounded-xl bg-purple-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">Websocket字节数<br/>
                    <span
                        className="font-semibold">{formatNumber(overview.WSBytes)}</span>
                </div>
            </div>
            {msg && msg.length > 0 && msg[0] &&
                msg.map((msg, i) => (
                    <div className={'h-[250px]'}>
                        <ResponsiveContainer>
                            <AreaChart
                                accessibilityLayer
                                data={msg}
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
                                <Tooltip content={({label, payload}) => (
                                    <div>
                                        {payload[0] &&
                                            <div
                                                className="flex h-auto min-w-[120px] items-center gap-x-2 rounded-medium bg-foreground p-2 text-tiny shadow-small">
                                                <div className="flex w-full flex-col gap-y-0">
                                                    <div className="flex w-full items-center gap-x-2">
                                                        <div
                                                            className="flex w-full items-center gap-x-1 text-small text-background">
                                                            <span>{payload[0].value}</span>
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>}
                                    </div>
                                )}/>
                                <YAxis
                                    axisLine={false}
                                    tickLine={false}
                                    tickMargin={8}
                                    tickFormatter={(value) => Math.round(value).toString()}
                                    tickCount={6}
                                    style={{fontSize: "var(--heroui-font-size-tiny)"}}
                                />
                                <Area
                                    type="monotone"
                                    dataKey="Count"
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
                                        onClick: (e, payload) => {
                                            console.log('点击了数据点', payload);
                                            // payload 包含了点击的数据信息
                                        }
                                    }}

                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>

                ))
            }
        </div>
    );
}

export default StatusPage;