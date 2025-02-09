import React, {useState} from 'react';
import ReactECharts from 'echarts-for-react';
import {AutoComplete} from "antd";
import axios from "axios";

const Charts = () => {
    const { dates, data } = generateMonthlyData();
    const eoptions = {
        grid: {top: 8, right: 8, bottom: 24, left: 36},
        xAxis: {
            type: 'category',
            data: dates,
        },
        yAxis: {
            type: 'value',
        },
        series: [
            {
                data: data,
                type: 'line',
                smooth: true,
            },
        ],
        tooltip: {
            trigger: 'axis',
        },
    };


    function generateMonthlyData() {
        const dates = [];
        const data = [];
        const now = new Date();

        for (let i = 29; i >= 0; i--) {
            const date = new Date();
            date.setDate(now.getDate() - i);

            const formattedDate = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
            dates.push(formattedDate);

            // 模拟数据：生成 0 到 100 的随机值
            data.push(Math.floor(Math.random() * 100));
        }
        return { dates, data };
    }

    const [value, setValue] = useState('');
    const [options, setOptions] = useState([]);
    const fetchKey = (t) => {
        axios.get(
            "http://localhost:8080/liver?key=" + t).then(res => {
                var array = []
            res.data.result.forEach((item, index) => {
                array.push({value: item})
            })
            setOptions(array)

        })
    }
    const onSelect = (data) => {
        console.log('onSelect', data);
    };
    const onChange = (data) => {
        setValue(data);
    };

    return (

        <div>
            <div className='row'>
                <AutoComplete
                    options={options}
                    style={{width: 200}}
                    onSelect={onSelect}
                    onSearch={(text) => fetchKey(text)}
                    placeholder="input here"
                />
                <AutoComplete
                    options={[
                        {value: '粉丝量'},
                        {value: 'lucy'},
                    ]}
                    style={{width: 200}}
                    onSelect={onSelect}
                    placeholder="input here"
                />
            </div>
            <ReactECharts option={eoptions}/>

        </div>


    )

}
export default Charts;