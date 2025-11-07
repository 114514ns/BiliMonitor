import china from '../cn.json';
import * as echarts from 'echarts'; //全局引入 ，可按需引入
import React, { useEffect, useRef , useState } from 'react';
import axios from "axios";
import {Modal, ModalHeader,ModalBody,ModalFooter,Button,ModalContent} from "@heroui/react";
import {NavLink} from "react-router-dom";
const GeoElement = (props) => {
    const chartRef = useRef();
    const topNumber = props.data[0].value;
    const bottomNumber = props.data[props.data.length - 1].value;
    const minValue = Math.min(...props.data.map(d => d.value));
    const maxValue = Math.max(...props.data.map(d => d.value));
    const echartsMapClick = (a,b) => {
        props.onClick(a.name)
    };

    const mapOption = (mapName, data) => {
        const myChart = echarts.init(chartRef.current);

        echarts.registerMap(mapName, data);
        const option = {
            tooltip: {
                backgroundColor: 'rgba(21, 24, 45, 0.9)', // 提示框浮层的背景颜色。
                textStyle: {
                    // 提示框浮层的文本样式。
                    color: '#fff',
                    fontSize: 14,
                },
                extraCssText: 'border-color: rgba(21, 24, 45, 0.9);',
            },
            visualMap: {
                min: minValue,
                max: maxValue,
                left: 'left',
                top: 'bottom',
                text: [topNumber, bottomNumber],
                inRange: {
                    color: ['#99ccff', '#003399'],
                },
                show: true, //图注
            },
            geo: {
                map: 'china',
                roam: false, //不开启缩放和平移
                zoom: 1.23, //视角缩放比例
                label: {
                    normal: {
                        show: true,
                        fontSize: '10',
                        color: 'rgba(0,0,0,0.7)',
                    },
                },
                itemStyle: {
                    normal: {
                        borderColor: 'rgba(0, 0, 0, 0.2)',
                    },
                    emphasis: {
                        areaColor: '#4BD6C7', //鼠标选择区域颜色
                        shadowOffsetX: 0,
                        shadowOffsetY: 0,
                        shadowBlur: 20,
                        borderWidth: 0,
                        shadowColor: 'rgba(0, 0, 0, 0.5)',
                    },
                },
            },
            series: [
                {
                    name:'Livers',
                    type: 'map',
                    geoIndex: 0,
                    data: props.data,
                },
            ],
        };
        myChart.setOption(option); //绘图
        //点击画布内还是画布外
        myChart.getZr().on('click', (params) => {
            if (params.target) {
                myChart.on('click', echartsMapClick); //增加点击事件
            }
        });
    };
    const loadingChina = () => {
        mapOption('china', china); //初始化-创建中国地图
    };

    useEffect(() => {
        loadingChina();
    }, [props.data]);

    return <div style={{ width: '100%', minHeight: '700px' }} ref={chartRef} />;
};



const GeoPage = (props) => {
    const [data,setData] = useState([])
    const [show,setShow] = useState(false)
    const [province,setProvince] = useState('')
    const [livers,setLivers] = useState([])
    useEffect(() => {
        axios.get("/api/api/geo").then((res) => {
            var tmp = res.data.data
            var objects = []
            Object.keys(tmp).forEach(key => {
                objects.push({
                    name: key,
                    value: tmp[key]
                })
            })
            setData(objects)
        })
    },[])
    useEffect(() => {
        axios.get("/api/api/geo/province?name=" + province).then((res) => {
            setLivers(res.data.data)
        })
    },[province])
    return (
        <div>
            {show &&       <Modal isOpen={show} onOpenChange={() => {
                setShow(!show)
            }} scrollBehavior={'inside'}>
                <ModalContent>
                    <>
                        <ModalHeader className="flex flex-col gap-1">{province}</ModalHeader>
                        <ModalBody>
                            <div className="flex flex-row flex-wrap">
                                {livers.map((item, i) => {
                                    return (
                                        <NavLink
                                            key={i}
                                            className="flex flex-col items-center w-20 m-2"
                                            to={'/liver/' + item.UID}
                                        >
                                            <img
                                                src={`${AVATAR_API}${item.UID}`}
                                                className="w-12 h-12"
                                                style={{ borderRadius: '50%' }}
                                            />
                                            <p
                                                className="text-center text-sm truncate w-full"
                                                title={item.Name}
                                            >
                                                {item.Name}
                                            </p>
                                        </NavLink>
                                    );
                                })}
                            </div>
                        </ModalBody>
                        <ModalFooter>
                            <Button color="danger" variant="light" onPress={() => {
                                setShow(false)
                            }}>
                                Close
                            </Button>
                        </ModalFooter>
                    </>
                </ModalContent>
            </Modal>}
            {data.length && <GeoElement data={data} onClick={(e) => {
                setShow(true)
                setProvince(e)
            }}/>}
        </div>
    )
}
export default GeoPage